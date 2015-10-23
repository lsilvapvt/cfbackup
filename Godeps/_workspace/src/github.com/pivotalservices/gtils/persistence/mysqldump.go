package persistence

import (
	"fmt"
	"io"

	"github.com/pivotalservices/gtils/command"
	"github.com/pivotalservices/gtils/osutils"
	"github.com/xchapter7x/lo"
)

const (
	MSQLDMP_CONNECT_CMD string = "%s -u %s -h %s --password=%s"
	MSQLDMP_CREATE_CMD         = "%s < %s"
	MSQLDMP_FLUSH_CMD          = "%s > flush privileges"
	MSQLDMP_DUMP_CMD           = "%s --all-databases"
)

var (
	MSQLDMP_DUMP_BIN string = "/var/vcap/packages/mariadb/bin/mysqldump"
	MSQLDMP_SQL_BIN         = "/var/vcap/packages/mariadb/bin/mysql"
)

type MysqlDump struct {
	Ip         string
	Username   string
	Password   string
	DbFile     string
	ConfigFile string
	Caller     command.Executer
	RemoteOps  remoteOperationsInterface
}

func NewMysqlDump(ip, username, password string) *MysqlDump {
	lo.G.Debug("setting up a new local mysqldump object")
	m := &MysqlDump{
		Ip:       ip,
		Username: username,
		Password: password,
		Caller:   command.NewLocalExecuter(),
	}
	return m
}

func NewRemoteMysqlDump(username, password string, sshCfg command.SshConfig) (*MysqlDump, error) {
	lo.G.Debug("setting up a new remote MyslDump object")
	remoteExecuter, err := command.NewRemoteExecutor(sshCfg)

	return &MysqlDump{
		Ip:        "localhost",
		Username:  username,
		Password:  password,
		Caller:    remoteExecuter,
		RemoteOps: osutils.NewRemoteOperations(sshCfg),
	}, err
}

func (s *MysqlDump) Import(lfile io.Reader) (err error) {
	if err = s.RemoteOps.UploadFile(lfile); err == nil {
		err = s.restore()
	}
	lo.G.Debug("mysqldump Import called: ", err)
	return
}

func (s *MysqlDump) Dump(dest io.Writer) (err error) {
	err = s.Caller.Execute(dest, s.getDumpCommand())
	lo.G.Debug("mysqldump Dump called: ", s.getDumpCommand(), err)
	return
}

func (s *MysqlDump) restore() (err error) {
	callList := []string{
		s.getImportCommand(),
		s.getFlushCommand(),
	}
	err = execute_list(callList, s.Caller)
	lo.G.Debug("mysqldump restore called: ", callList, err)
	return
}

func (s *MysqlDump) getImportCommand() string {
	return fmt.Sprintf(MSQLDMP_CREATE_CMD, s.getConnectCommand(MSQLDMP_SQL_BIN), s.RemoteOps.Path())
}

func (s *MysqlDump) getFlushCommand() string {
	return fmt.Sprintf(MSQLDMP_FLUSH_CMD, s.getConnectCommand(MSQLDMP_SQL_BIN))
}

func (s *MysqlDump) getDumpCommand() string {
	return fmt.Sprintf(MSQLDMP_DUMP_CMD, s.getConnectCommand(MSQLDMP_DUMP_BIN))
}

func (s *MysqlDump) getConnectCommand(bin string) string {
	return fmt.Sprintf(MSQLDMP_CONNECT_CMD,
		bin,
		s.Username,
		s.Ip,
		s.Password,
	)
}
