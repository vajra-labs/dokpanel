package service

import (
	"goploy/src/conf"
	"goploy/src/db/repos"
	"goploy/src/pkg/docker"
)

type ScheduleService struct {
	cfg   *conf.Config
	query *repos.Queries
	paths *docker.AppPaths
}

func NewSchedule(cfg *conf.Config, query *repos.Queries, paths *docker.AppPaths) *ScheduleService {
	return &ScheduleService{cfg: cfg, query: query, paths: paths}
}
