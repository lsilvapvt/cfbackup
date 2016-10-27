package opsmanager

import (
	"github.com/pivotalservices/cfbackup"
	"github.com/pivotalservices/cfbackup/tileregistry"
	"github.com/xchapter7x/lo"
)

//New -- builds a new ops manager object pre initialized
func (s *OpsManagerBuilder) New(tileSpec tileregistry.TileSpec) (opsManagerTileCloser tileregistry.TileCloser, err error) {
	var opsManager *OpsManager
	opsManager, err = NewOpsManager(tileSpec.OpsManagerHost, tileSpec.AdminUser, tileSpec.AdminPass, tileSpec.AdminToken, tileSpec.OpsManagerUser, tileSpec.OpsManagerPass, tileSpec.ClientID, tileSpec.ClientSecret, tileSpec.OpsManagerPassphrase, tileSpec.ArchiveDirectory, tileSpec.CryptKey)
	opsManager.ClearBoshManifest = tileSpec.ClearBoshManifest

	if installationSettings, err := opsManager.GetInstallationSettings(); err == nil {
		config := cfbackup.NewConfigurationParserFromReader(installationSettings)

		if iaas, hasKey := config.GetIaaS(); hasKey {
			lo.G.Debug("we found a iaas info block")
			opsManager.SetSSHPrivateKey(iaas.SSHPrivateKey)

		} else {
			lo.G.Debug("No IaaS PEM key found. Defaulting to using ssh username and password credentials")
		}
	}
	opsManagerTileCloser = struct {
		tileregistry.Tile
		tileregistry.Closer
	}{
		opsManager,
		new(tileregistry.DoNothingCloser),
	}
	return
}
