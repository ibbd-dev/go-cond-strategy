package servicesDegrade

import (
	"sync"
	"time"

	"github.com/ibbd-dev/go-tools/timer"
)

const (
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
	OneMinute TMetric

	// 5分钟指标
	FiveMinute TMetric
}

// 统计指标
type TMetric struct {
	Count   uint32 // 时间段内事件发生的次数
	AvgTime int64  // 时间段内事件消耗的平均时间(总耗时 / 总次数)
}

var (
	// 每分钟执行一次
	// 方便测试
	updateMetricDuration = time.Minute
	updateMetricFunc     = updateMetric

	metricMu     sync.Mutex
	eventsMetric *TEventsMetric
)

func init() {
	eventsMetric = &TEventsMetric{
		Events: make(map[string]*TEventMetric),
	}
	timer.AddFunc(updateMetricFunc, updateMetricDuration)
}

// 更新指标
func updateMetric() {
	println("====================")
	metricMu.Lock()
	eventsRW.Lock()

	now := nowFunc().UnixNano()
	for name, ev := range events {
		eventsMetric.Events[name].OneMinute = calMetric(now, &ev.oneMinute)
		ev.oneMinute.count = 0
		ev.oneMinute.totalTime = 0
		ev.oneMinute.beginTime = now

		if now-ev.fiveMinute.beginTime > fiveMinute {
			eventsMetric.Events[name].FiveMinute = calMetric(now, &ev.fiveMinute)
			ev.fiveMinute.count = 0
			ev.fiveMinute.totalTime = 0
			ev.fiveMinute.beginTime = now
		}
	}
	eventsRW.Unlock()

	// 判断降级配置
	confList.Lock()
	for _, conf := range confList.conf {
		if conf.Check(eventsMetric) {
			conf.YesAction()
		} else {
			conf.NoAction()
		}
	}
	confList.Unlock()

	metricMu.Unlock()
}

// 计算事件的指标
func calMetric(now int64, data *tStatData) (metric TMetric) {
	if data.count > 0 {
		diff := now - data.beginTime
		diffRate := float64(updateMetricDuration) / float64(diff)

		// 总数需要对消耗的时间做一次平滑
		// 例如统计周期本来设置为1分钟，但是实际上跑了2分钟，2分钟内count=100，那么对于到1分钟应该是count=50
		metric.Count = uint32(float64(data.count) * diffRate)
		metric.AvgTime = data.totalTime / int64(data.count)
	}

	return metric
}
