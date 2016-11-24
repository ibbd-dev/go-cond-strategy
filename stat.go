package servicesDegrade

import (
	"sync"
	"time"

	"github.com/ibbd-dev/go-tools/timer"
)

const (
	// 每分钟执行一次
	duration = time.Minute

	second     int64 = 1000 * 1000 * 1000 // 单位：纳秒
	minute     int64 = 60 * second
	fiveMinute int64 = 5 * minute
)

// 所有事件的指标
type TEventsMetric struct {
	Events map[string]*TEventMetric
}

// 单个事件的指标
type TEventMetric struct {
	// 1分钟指标
	OneMinute *TMetric

	// 5分钟指标
	FiveMinute *TMetric
}

// 统计指标
type TMetric struct {
	Count   uint32 // 时间段内事件发生的次数
	AvgTime int64  // 时间段内事件消耗的平均时间(总耗时 / 总次数)
}

var (
	metricMu     sync.Mutex
	eventsMetric *TEventsMetric
)

func init() {
	timer.AddFunc(updateMetric, duration)
}

func updateMetric() {
	metricMu.Lock()
	eventsRW.Lock()

	now := nowFunc().UnixNano()
	for name, ev := range events {
		eventsMetric.Events[name].OneMinute = calMetric(now, ev.oneMinute)
		ev.oneMinute.count = 0
		ev.oneMinute.totalTime = 0
		ev.oneMinute.beginTime = now

		if now-ev.fiveMinute.beginTime > fiveMinute {
			eventsMetric.Events[name].FiveMinute = calMetric(now, ev.fiveMinute)
			ev.fiveMinute.count = 0
			ev.fiveMinute.totalTime = 0
			ev.fiveMinute.beginTime = now
		}
	}
	eventsRW.Unlock()

	// 判断降级配置
	confList.Lock()
	for _, conf := range confList.Conf {
		if conf.Check(eventsMetric) {
			conf.YesAction()
		} else {
			conf.NoAction()
		}
	}
	confList.Unlock()

	metricMu.Unlock()
}

func calMetric(now int64, data *tStatData) *TMetric {
	diff := now - data.beginTime
	diffRate := float64(now) / float64(diff)

	return &TMetric{
		Count:   uint32(float64(data.count) * diffRate),
		AvgTime: data.totalTime / int64(data.count),
	}
}
