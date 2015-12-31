package cfbackup

import "path"

const (
	BACKUP_LOGGER_NAME  = "Backup"
	RESTORE_LOGGER_NAME = "Restore"
)

var (
	TILE_RESTORE_ACTION = func(t Tile) func() error {
		return t.Restore
	}
	TILE_BACKUP_ACTION = func(t Tile) func() error {
		return t.Backup
	}
)

//Backup the list of all default tiles
func RunBackupPipeline(hostname, adminUsername, adminPassword, opsManagerUsername, opsManagerPassword, destination string) (err error) {
	var tiles []Tile
	conn := connectionBucket{
		hostname:           hostname,
		adminUsername:      adminUsername,
		adminPassword:      adminPassword,
		opsManagerUsername: opsManagerUsername,
		opsManagerPassword: opsManagerPassword,
		destination:        destination,
	}

	if tiles, err = fullTileList(conn, BACKUP_LOGGER_NAME); err == nil {
		err = RunPipeline(TILE_BACKUP_ACTION, tiles)
	}
	return
}

//Restore the list of all default tiles
func RunRestorePipeline(hostname, adminUsername, adminPassword, opsManagerUser, opsManagerPassword, destination string) (err error) {
	var tiles []Tile
	conn := connectionBucket{
		hostname:           hostname,
		adminUsername:      adminUsername,
		adminPassword:      adminPassword,
		opsManagerUsername: opsManagerUser,
		opsManagerPassword: opsManagerPassword,
		destination:        destination,
	}

	if tiles, err = fullTileList(conn, RESTORE_LOGGER_NAME); err == nil {
		err = RunPipeline(TILE_RESTORE_ACTION, tiles)
	}
	return
}

//Runs a pipeline action (restore/backup) on a list of tiles
var RunPipeline = func(actionBuilder func(Tile) func() error, tiles []Tile) (err error) {
	var pipeline []action

	for _, tile := range tiles {
		pipeline = append(pipeline, actionBuilder(tile))
	}
	err = runActions(pipeline)
	return
}

func runActions(actions []action) (err error) {
	for _, action := range actions {

		if err = action(); err != nil {
			break
		}
	}
	return
}

func fullTileList(conn connBucketInterface, loggerName string) (tiles []Tile, err error) {
	var (
		opsmanager     Tile
		elasticRuntime Tile
	)
	installationFilePath := path.Join(conn.Destination(), OPSMGR_BACKUP_DIR, OPSMGR_INSTALLATION_SETTINGS_FILENAME)

	if opsmanager, err = NewOpsManager(conn.Host(), conn.AdminUser(), conn.AdminPass(), conn.OpsManagerUser(), conn.OpsManagerPass(), conn.Destination()); err == nil {
		elasticRuntime = NewElasticRuntime(installationFilePath, conn.Destination())
		tiles = []Tile{
			opsmanager,
			elasticRuntime,
		}
	}
	return
}
