package cronmanager

import (
	"fmt"
	"shared/repository"
	"sync"

	"github.com/robfig/cron/v3"
)

var scheduler = cron.New()
var (
	entryMap   = make(map[string]cron.EntryID)
	entryMutex sync.Mutex
)

var backupFunc func(name string) error

func init() {
	scheduler.Start()
}

// RegisterBackupFunc registers the function used to perform backup operations.
// This allows high-level backup logic to be injected, avoiding circular dependencies.
func RegisterBackupFunc(f func(name string) error) {
	backupFunc = f
}

// LoadCrons loads all active cron jobs from the database and starts scheduling them.
func LoadCrons() error {
	db := repository.GetDB()
	var cronJobs []repository.Cron
	err := db.Where("active = ?", true).Find(&cronJobs).Error
	if err != nil {
		return err
	}

	for _, cronJob := range cronJobs {
		entryID, err := scheduler.AddFunc(cronJob.Schedule, makeCronFunc(cronJob.Name))
		if err != nil {
			fmt.Printf("Error scheduling cron for database %s: %v\n", cronJob.Name, err)
		} else {
			entryMutex.Lock()
			entryMap[cronJob.Name+":"+cronJob.Schedule] = entryID
			entryMutex.Unlock()
		}
	}

	return nil
}

func makeCronFunc(name string) func() {
	return func() {
		if backupFunc != nil {
			if err := backupFunc(name); err != nil {
				fmt.Printf("Error performing backup for database %s: %v\n", name, err)
			}
		} else {
			fmt.Println("Warning: Backup function not registered in cronmanager")
		}
	}
}

// AddCronJob schedules a new cron job and registers it in the active job tracking map.
func AddCronJob(name, schedule string) error {
	entryMutex.Lock()
	defer entryMutex.Unlock()

	key := name + ":" + schedule
	if _, exists := entryMap[key]; exists {
		return nil // Already scheduled
	}

	entryID, err := scheduler.AddFunc(schedule, makeCronFunc(name))
	if err != nil {
		return err
	}

	entryMap[key] = entryID
	return nil
}

// RemoveCronJob unschedules an active cron job and removes it from the tracking map.
func RemoveCronJob(name, schedule string) {
	entryMutex.Lock()
	defer entryMutex.Unlock()

	key := name + ":" + schedule
	if entryID, exists := entryMap[key]; exists {
		scheduler.Remove(entryID)
		delete(entryMap, key)
	}
}
