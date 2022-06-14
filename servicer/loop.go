package servicer

import "time"

// Start event cycle
func (s *Servicer) loop() {

	// Route changes are updated every 8 hours
	loadUpdateFileTicker := time.NewTicker(time.Hour * 8)
	checkCustomerActiveTicker := time.NewTicker(time.Second * 35)

	for {
		select {
		case <-loadUpdateFileTicker.C:
			// Automatically update routes
			s.LoadRoutesUpdate()
		case <-checkCustomerActiveTicker.C:
			// Check client heartbeat
			s.checkCustomerActive()
		}
	}
}
