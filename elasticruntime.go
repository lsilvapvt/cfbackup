package cfbackup

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudfoundry-community/go-cfenv"
	ghttp "github.com/pivotalservices/gtils/http"
	"github.com/pivotalservices/gtils/log"
	"github.com/xchapter7x/lo"
)

// NewElasticRuntime initializes an ElasticRuntime intance
var NewElasticRuntime = func(jsonFile string, target string, sshKey string) *ElasticRuntime {
	systemsInfo := NewSystemsInfo(jsonFile, sshKey)
	context := &ElasticRuntime{
		SSHPrivateKey:     sshKey,
		JSONFile:          jsonFile,
		BackupContext:     NewBackupContext(target, cfenv.CurrentEnv()),
		SystemsInfo:       systemsInfo.SystemDumps,
		PersistentSystems: systemsInfo.PersistentSystems(),
	}
	return context
}

// Backup performs a backup of a Pivotal Elastic Runtime deployment
func (context *ElasticRuntime) Backup() (err error) {
	return context.backupRestore(ExportArchive)
}

// Restore performs a restore of a Pivotal Elastic Runtime deployment
func (context *ElasticRuntime) Restore() (err error) {
	err = context.backupRestore(ImportArchive)
	return
}

func (context *ElasticRuntime) backupRestore(action int) (err error) {
	var (
		ccJobs []CCJob
	)

	if err = context.ReadAllUserCredentials(); err == nil && context.directorCredentialsValid() {
		lo.G.Debug("Retrieving All CC VMs")
		manifest, erro := context.getManifest()
		if err != nil {
			return erro
		}
		if ccJobs, err = context.getAllCloudControllerVMs(); err == nil {
			directorInfo := context.SystemsInfo[ERDirector]
			cloudController := NewCloudController(directorInfo.Get(SDIP), directorInfo.Get(SDUser), directorInfo.Get(SDPass), context.InstallationName, manifest, ccJobs)
			lo.G.Debug("Setting up CC jobs")
			defer cloudController.Start()
			cloudController.Stop()
		}
		lo.G.Debug("Running db action")
		if len(context.PersistentSystems) > 0 {
			err = context.RunDbAction(context.PersistentSystems, action)
			if err != nil {
				lo.G.Error("Error backing up db", err)
				err = ErrERDBBackup
			}
		} else {
			err = ErrEREmptyDBList
		}
	} else if err == nil {
		err = ErrERDirectorCreds
	}
	return
}

func (context *ElasticRuntime) getAllCloudControllerVMs() (ccvms []CCJob, err error) {

	lo.G.Debug("Entering getAllCloudControllerVMs() function")
	directorInfo := context.SystemsInfo[ERDirector]
	connectionURL := fmt.Sprintf(ERVmsURL, directorInfo.Get(SDIP), context.InstallationName)
	lo.G.Debug("getAllCloudControllerVMs() function", log.Data{"connectionURL": connectionURL, "directorInfo": directorInfo})
	gateway := context.HTTPGateway
	if gateway == nil {
		gateway = ghttp.NewHttpGateway()
	}
	lo.G.Debug("Retrieving CC vms")
	if resp, err := gateway.Get(ghttp.HttpRequestEntity{
		Url:         connectionURL,
		Username:    directorInfo.Get(SDUser),
		Password:    directorInfo.Get(SDPass),
		ContentType: "application/json",
	})(); err == nil {
		var jsonObj []VMObject

		lo.G.Debug("Unmarshalling CC vms")
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err = json.Unmarshal(body, &jsonObj); err == nil {
			ccvms, err = GetCCVMs(jsonObj)
		}
	}
	return
}

//RunDbAction - run a db action dump/import against a list of systemdump types
func (context *ElasticRuntime) RunDbAction(dbInfoList []SystemDump, action int) (err error) {

	for _, info := range dbInfoList {
		lo.G.Debug(fmt.Sprintf("RunDbAction info: %+v", info))

		if err = info.Error(); err == nil {
			err = context.readWriterArchive(info, context.TargetDir, action)
		} else {
			// Don't error out yet until issue #111461510 is resolved.
			continue
		}
	}
	return
}

func (context *ElasticRuntime) readWriterArchive(dbInfo SystemDump, databaseDir string, action int) (err error) {
	filename := fmt.Sprintf(ERBackupFileFormat, dbInfo.Get(SDComponent))
	filepath := path.Join(databaseDir, filename)

	var pb PersistanceBackup

	if pb, err = dbInfo.GetPersistanceBackup(); err == nil {
		switch action {
		case ImportArchive:
			lo.G.Debug("Restoring %s", dbInfo.Get(SDComponent))
			var backupReader io.ReadCloser
			if backupReader, err = context.Reader(filepath); err == nil {
				defer backupReader.Close()
				err = pb.Import(backupReader)
				lo.G.Debug("Done restoring %s", dbInfo.Get(SDComponent))
			}
		case ExportArchive:
			lo.G.Info("Exporting %s", dbInfo.Get(SDComponent))
			var backupWriter io.WriteCloser
			if backupWriter, err = context.Writer(filepath); err == nil {
				defer backupWriter.Close()
				err = pb.Dump(backupWriter)
				lo.G.Debug("Done backing up %s", dbInfo.Get(SDComponent))
			}
		}
	}
	return
}

//ReadAllUserCredentials - get all user creds from the installation json
func (context *ElasticRuntime) ReadAllUserCredentials() (err error) {
	var (
		fileRef *os.File
		jsonObj InstallationCompareObject
	)
	defer fileRef.Close()

	if fileRef, err = os.Open(context.JSONFile); err == nil {
		if jsonObj, err = ReadAndUnmarshal(fileRef); err == nil {
			err = context.assignCredentialsAndInstallationName(jsonObj)
		}
	}
	return
}

func (context *ElasticRuntime) assignCredentialsAndInstallationName(jsonObj InstallationCompareObject) (err error) {

	if err = context.assignCredentials(jsonObj); err == nil {
		context.InstallationName, err = GetDeploymentName(jsonObj)
	}
	return
}

func (context *ElasticRuntime) assignCredentials(jsonObj InstallationCompareObject) (err error) {

	for name, sysInfo := range context.SystemsInfo {
		var (
			ip    string
			pass  string
			vpass string
		)
		sysInfo.Set(SDVcapUser, ERDefaultSystemUser)
		sysInfo.Set(SDUser, sysInfo.Get(SDIdentity))

		if ip, pass, err = GetPasswordAndIP(jsonObj, sysInfo.Get(SDProduct), sysInfo.Get(SDComponent), sysInfo.Get(SDIdentity)); err == nil {
			sysInfo.Set(SDIP, ip)
			sysInfo.Set(SDPass, pass)
			lo.G.Debug("%s credentials for %s from installation.json are %s", name, sysInfo.Get(SDComponent), sysInfo.Get(SDIdentity), pass)
			_, vpass, err = GetPasswordAndIP(jsonObj, sysInfo.Get(SDProduct), sysInfo.Get(SDComponent), sysInfo.Get(SDVcapUser))
			sysInfo.Set(SDVcapPass, vpass)
			context.SystemsInfo[name] = sysInfo
		}
	}
	return
}

func (context *ElasticRuntime) directorCredentialsValid() (ok bool) {
	var directorInfo SystemDump

	if directorInfo, ok = context.SystemsInfo[ERDirector]; ok {
		connectionURL := fmt.Sprintf(ERDirectorInfoURL, directorInfo.Get(SDIP))
		gateway := context.HTTPGateway
		if gateway == nil {
			gateway = ghttp.NewHttpGateway()
		}
		_, err := gateway.Get(ghttp.HttpRequestEntity{
			Url:         connectionURL,
			Username:    directorInfo.Get(SDUser),
			Password:    directorInfo.Get(SDPass),
			ContentType: "application/json",
		})()
		ok = (err == nil)
	}
	return
}

func (context *ElasticRuntime) getManifest() (manifest string, err error) {
	directorInfo, _ := context.SystemsInfo[ERDirector]
	director := NewDirector(directorInfo.Get(SDIP), directorInfo.Get(SDUser), directorInfo.Get(SDPass), 25555)
	mfs, err := director.GetDeploymentManifest(context.InstallationName)
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(mfs)
	if err != nil {
		return
	}
	return string(data), nil
}
