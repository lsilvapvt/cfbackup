package cfbackup_test

import (
	"github.com/cloudfoundry-community/go-cfenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var _ = Describe("BackupContext", func() {
	Describe("given a NewBackupContext() func", func() {
		Context("when called with a targetdir", func() {
			var backupContext BackupContext
			var controlTargetDir = "random/path/to/archive"
			BeforeEach(func() {
				backupContext = NewBackupContext(controlTargetDir, cfenv.CurrentEnv())
			})
			It("then it should create a backup context with the targetdir set", func() {
				Ω(backupContext.TargetDir).Should(Equal(controlTargetDir))
				Ω(backupContext.AccessKeyID).Should(BeEmpty())
				Ω(backupContext.SecretAccessKey).Should(BeEmpty())
				Ω(backupContext.BucketName).Should(BeEmpty())
				Ω(backupContext.IsS3Archive).Should(BeFalse())
			})
		})

		Context("when called with a complete set of s3 information", func() {
			var backupContext BackupContext
			var controlTargetDir = "random/path/to/archive"
			var controlkey = "accesskeyid"
			var controlSecret = "secretkey"
			var controlBucket = "bucketname"
			var controlS3Active = "true"
			BeforeEach(func() {
				backupContext = NewBackupContext(controlTargetDir, map[string]string{
					AccessKeyIDVarname:     controlkey,
					SecretAccessKeyVarname: controlSecret,
					BucketNameVarname:      controlBucket,
					IsS3Varname:            controlS3Active,
				})
			})
			It("then it should create a backup context that can be used for s3 backup/restore ", func() {
				Ω(backupContext.TargetDir).Should(Equal(controlTargetDir))
				Ω(backupContext.AccessKeyID).Should(Equal(controlkey))
				Ω(backupContext.SecretAccessKey).Should(Equal(controlSecret))
				Ω(backupContext.BucketName).Should(Equal(controlBucket))
				Ω(backupContext.IsS3Archive).Should(BeTrue())
			})
		})
	})
})
