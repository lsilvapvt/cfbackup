package cfbackup

import "fmt"

// GetIPsByProductAndJob finds a product and jobName
func (s *InstallationSettings) GetIPsByProductAndJob(productName string, jobName string) (ips []string) {

	if s.isLegacyFormat() {

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
func (s *InstallationSettings) FindJobsByProductID(id string) ([]Jobs) {
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
    return true
}
