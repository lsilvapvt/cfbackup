package cfbackup

import "fmt"

//FindVMCredentialsByProductAndJob gets VMCredentials for a given product and job
func (s *InstallationSettings) FindVMCredentialsByProductAndJob(productName, jobName string) (vmCredentials VMCredentials, err error) {
    var product Products
    if product, err = s.FindByProductID(productName); err == nil {
        vmCredentials, err = product.GetVMCredentialsByJob(jobName)
    }
    return 
}
// FindIPsByProductAndJob finds a product and jobName
func (s *InstallationSettings) FindIPsByProductAndJob(productName string, jobName string) (IPs []string, err error) {

	if s.isLegacyFormat() {
		IPs, err = s.extractLegacyIPsForProductAndJob(productName, jobName)
	} else {
		IPs, err = s.extractIPsForProductAndJob(productName, jobName)
	}
	return
}

func (s *InstallationSettings) extractLegacyIPsForProductAndJob(productName, jobName string) (IPs []string, err error) {
	var product Products
	if product, err = s.FindByProductID(productName); err == nil {
		IPs = product.GetIPsByJob(jobName)
	}
	return
}

func (s *InstallationSettings) extractIPsForProductAndJob(productName, jobName string) (IPs []string, err error) {
	var product Products
	if product, err = s.FindByProductID(productName); err == nil {
		var job Jobs
		if job, err = product.GetJob(jobName); err == nil {
			IPs, err = s.findIPs(product, job)
		}
	}
	return
}

func (s *InstallationSettings) findIPs(product Products, job Jobs) (IPs []string, err error) {
	var IPsResponse []string
	for _, azGUID := range product.AZReference {
		if IPsResponse, err = s.IPAssignments.FindIPsByProductGUIDAndJobGUIDAndAvailabilityZoneGUID(product.GUID, job.GUID, azGUID); err == nil {
			for _, ip := range IPsResponse {
				IPs = append(IPs, ip)
			}
		}
	}
	return
}

// FindByProductID finds a product by product id
func (s *InstallationSettings) FindByProductID(id string) (productResponse Products, err error) {
	var found bool
	for _, product := range s.Products {
		identifier := product.Identifier
		if identifier == id {
			productResponse = product
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("Product not found %s", id)
	}

	return
}

// FindJobsByProductID finds all the jobs in an installation by product id
func (s *InstallationSettings) FindJobsByProductID(id string) []Jobs {
	cfJobs := []Jobs{}

	for _, product := range s.Products {
		identifier := product.Identifier
		if identifier == id {
			for _, job := range product.Jobs {
				cfJobs = append(cfJobs, job)
			}
		}
	}
	return cfJobs
}

// FindCFPostgresJobs finds all the postgres jobs in the cf product
func (s *InstallationSettings) FindCFPostgresJobs() (jobs []Jobs) {

	jobsList := s.FindJobsByProductID("cf")
	for _, job := range jobsList {
		if isPostgres(job.Identifier, job.Instances) {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func isPostgres(job string, instances []Instances) bool {
	pgdbs := []string{"ccdb", "uaadb", "consoledb"}

	for _, pgdb := range pgdbs {
		if pgdb == job {
			for _, instances := range instances {
				val := instances.Value
				if val >= 1 {
					return true
				}
			}
		}
	}
	return false
}

func (s *InstallationSettings) isLegacyFormat() bool {
	return s.IPAssignments.Assignments == nil
}
