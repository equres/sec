package cache

import (
	"context"

	"github.com/equres/sec/pkg/config"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	Redis *redis.Client
}

const (
	SECCacheStats           string = "cache.SECCacheStats"
	SECTopFiveRecentFilings string = "cache.SECTopFiveRecentFilings"
	SECCIKsCount            string = "cache.SECCIKsCount"
	SECFilesCount           string = "cache.SECFilesCount"
	SECCompaniesCount       string = "cache.SECCompaniesCount"
	SECMonthsInYear         string = "cache.SECMonthsInYear"
	SECDaysInMonth          string = "cache.SECDaysInMonth"
	SECCompaniesInDay       string = "cache.SECCompaniesInDay"
	SECFilingsInDay         string = "cache.SECFilingsInDay"
	SECCompanySlugs         string = "cache.SECCompanySlugs"
	SECCompanySlugsHTML     string = "cache.SECCompanySlugsHTML"
	SECCompanyFilings       string = "cache.SECCompanyFilings"
	SECCompanyFilingsHTML   string = "cache.SECCompanyFilingsHTML"
	SECSICs                 string = "cache.SECSICs"
	SECCompaniesWithSIC     string = "cache.SECCompaniesWithSIC"
	SECHourlyDownloadStats  string = "cache.SECHourlyDownloadStats"
	SECHours                string = "cache.SECHours"
	SECDownloadDates        string = "cache.SECDownloadDates"
	SECCompanies            string = "cache.SECCompanies"
)

func NewCache(cfg *config.Config) Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisURL(),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	return Cache{
		Redis: rdb,
	}
}

func (c *Cache) Set(k, v string) error {
	err := c.Redis.Set(context.Background(), k, v, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) Get(k string) (string, error) {
	v, err := c.Redis.Get(context.Background(), k).Result()
	if err != nil {
		return "", err
	}

	return v, err
}

func (c *Cache) MustSet(k, v string) error {
	err := c.Redis.Set(context.Background(), k, v, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) MustGet(k string) (string, error) {
	v, err := c.Redis.Get(context.Background(), k).Result()
	if err != nil {
		return "", err
	}

	return v, err
}
