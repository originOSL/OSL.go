// name: cron
// description: Task scheduling utilities
// author: roturbot
// requires: time, strings

type Cron struct {
	jobs map[string]*CronJob
}

type CronJob struct {
	name      string
	schedule  string
	callback  func()
	nextRun   time.Time
	enabled   bool
	lastRun   time.Time
	runCount  int
}

func (c *Cron) create() *Cron {
	return &Cron{
		jobs: make(map[string]*CronJob),
	}
}

func (c *Cron) addJob(name any, schedule any, callback any) bool {
	nameStr := OSLtoString(name)
	scheduleStr := OSLtoString(schedule)
	
	if !c.validateSchedule(scheduleStr) {
		return false
	}

	job := &CronJob{
		name:     nameStr,
		schedule: scheduleStr,
		nextRun:  c.calculateNextRun(scheduleStr),
		enabled:  true,
	}

	c.jobs[nameStr] = job
	return true
}

func (c *Cron) removeJob(name any) bool {
	nameStr := OSLtoString(name)
	_, exists := c.jobs[nameStr]
	delete(c.jobs, nameStr)
	return exists
}

func (c *Cron) enableJob(name any) bool {
	nameStr := OSLtoString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return false
	}
	job.enabled = true
	return true
}

func (c *Cron) disableJob(name any) bool {
	nameStr := OSLToString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return false
	}
	job.enabled = false
	return true
}

func (c *Cron) runJob(name any) bool {
	nameStr := OSLtoString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return false
	}

	if job.enabled {
		OSLcallFunc(callback, nil, []any{})
		job.lastRun = time.Now()
		job.runCount++
		job.nextRun = c.calculateNextRun(job.schedule)
	}
	return true
}

func (c *Cron) runAll() map[string]any {
	results := make(map[string]any)
	
	for name, job := range c.jobs {
		startTime := time.Now()
		results[name] = map[string]any{
			"success": c.runJob(name),
			"duration": time.Since(startTime).Seconds(),
			"count":    job.runCount,
		}
	}
	
	return results
}

func (c *Cron) validateSchedule(schedule any) bool {
	scheduleStr := OSLtoString(schedule)

	parts := strings.Fields(scheduleStr)
	if len(parts) < 5 {
		return false
	}

	return true
}

func (c *Cron) calculateNextRun(schedule string) time.Time {
	now := time.Now()
	
	return now.Add(time.Hour * 1)
}

func (c *Cron) start() chan any {
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan any)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.checkJobs()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return done
}

func (c *Cron) stop(done chan any) {
	if done != nil {
		close(done)
	}
}

func (c *Cron) checkJobs() {
	now := time.Now()

	for _, job := range c.jobs {
		if job.enabled && now.After(job.nextRun) {
			c.runJob(job.name)
		}
	}
}

func (c *Cron) getJobs() []map[string]any {
	jobs := make([]map[string]any, len(c.jobs))
	
	i := 0
	for name, job := range c.jobs {
		jobs[i] = map[string]any{
			"name":     job.name,
			"schedule": job.schedule,
			"enabled":  job.enabled,
			"nextRun":  job.nextRun.Unix(),
			"lastRun":  job.lastRun.Unix(),
			"runCount": job.runCount,
		}
		i++
	}

	return jobs
}

func (c *Cron) getJob(name any) map[string]any {
	nameStr := OSLtoString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return nil
	}

	return map[string]any{
		"name":     job.name,
		"schedule": job.schedule,
		"enabled":  job.enabled,
		"nextRun":  job.nextRun.Unix(),
		"lastRun":  job.lastRun.Unix(),
		"runCount": job.runCount,
	}
}

func (c *Cron) getJobCount() int {
	return len(c.jobs)
}

func (c *Cron) isEnabled(name any) bool {
	nameStr := OSLToString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return false
	}
	return job.enabled
}

func (c *Cron) getLastRun(name any) int64 {
	nameStr := OSLtoString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return 0
	}
	return job.lastRun.Unix()
}

func (c *Cron) getNextRun(name any) int64 {
	nameStr := OSLToString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return 0
	}
	return job.nextRun.Unix()
}

func (c *Cron) getRunCount(name any) int {
	nameStr := OSLtoString(name)
	job, exists := c.jobs[nameStr]
	if !exists {
		return 0
	}
	return job.runCount
}

func (c *Cron) runOnce() map[string]any {
	return c.runAll()
}

func (c *Cron) clear() {
	c.jobs = make(map[string]*CronJob)
}

var cron = Cron{}
