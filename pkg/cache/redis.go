package cache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type SecStatDownload struct {
	YearMonth      string
	Day            int
	DownloadOK     int
	DownloadFailed int
}

func CreateRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 60 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", "localhost:6379") },
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func UpdateDownloadStat(pool *redis.Pool, dateString string, isSuccesful bool) error {
	conn := pool.Get()
	defer conn.Close()

	if dateString == "" {
		return nil
	}

	date, err := time.Parse("01/02/2006", dateString)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("sec_download_%v%v%v", date.Year(), int(date.Month()), date.Day())

	redisStats, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		return err
	}

	stats := SecStatDownload{}
	err = redis.ScanStruct(redisStats, &stats)
	if err != nil {
		return err
	}

	stats.YearMonth = fmt.Sprintf("%v/%v", date.Year(), int(date.Month()))
	stats.Day = date.Day()
	if isSuccesful {
		stats.DownloadOK++
	} else {
		stats.DownloadFailed++
	}

	_, err = conn.Do("HSET", redis.Args{}.Add(key).AddFlat(stats)...)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", key, 10000000)
	if err != nil {
		return err
	}

	return nil
}

func GetAllStats(pool *redis.Pool) ([]SecStatDownload, error) {
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		return nil, err
	}

	stats := []SecStatDownload{}

	for _, key := range keys {
		redisStats, err := redis.Values(conn.Do("HGETALL", key))
		if err != nil {
			return nil, err
		}

		val := SecStatDownload{}
		err = redis.ScanStruct(redisStats, &val)
		if err != nil {
			return nil, err
		}

		stats = append(stats, val)
	}

	return stats, nil
}
