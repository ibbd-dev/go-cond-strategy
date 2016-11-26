package crondStrategy

import (
	"sync"
	"time"

	"github.com/ibbd-dev/go-tools/timer"
)

const (
	second     = int64(time.Second) // 单位：纳秒
	minute     = int64(time.Minute)
	fiveMinute = 5 * minute
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
	now := nowFunc().UnixNano()
	//println("====================")

	metricMu.Lock()
	eventsRW.Lock()
	for name, ev := range events {
		// 將1分钟的统计数据加到5分钟的统计数据上
		ev.fiveMinute.count += ev.oneMinute.count
		ev.fiveMinute.totalTime += ev.oneMinute.totalTime

		// 计算1分钟的指标
		eventsMetric.Events[name].OneMinute = calMetric(now, &ev.oneMinute)
		ev.oneMinute.count = 0
		ev.oneMinute.totalTime = 0
		ev.oneMinute.beginTime = now

		if now-ev.fiveMinute.beginTime > fiveMinute {
			// 计算5分钟的指标
			eventsMetric.Events[name].FiveMinute = calMetric(now, &ev.fiveMinute)
			ev.fiveMinute.count = 0
			ev.fiveMinute.totalTime = 0
			ev.fiveMinute.beginTime = now
		}
	}
	eventsRW.Unlock()

	// 普通的策略配置
	confList.Lock()
	for _, conf := range confList.conf {
		if conf.Check(eventsMetric) {
			conf.YesAction()
		} else {
			conf.NoAction()
		}
	}
	confList.Unlock()

	// 降级服务配置
	degConfList.Lock()
	for _, conf := range degConfList.conf {
		level := conf.Check(eventsMetric)
		if level != conf.lastLevel {
			conf.lastLevel = parseLevel(conf.lastLevel, level, conf)
		}
	}
	degConfList.Unlock()

	metricMu.Unlock()
}

// 计算事件的指标
func calMetric(now int64, data *tStatData) (metric TMetric) {
	if data.count > 0 {
		diffRate := float64(updateMetricDuration) / float64(now-data.beginTime)

		// 总数需要对消耗的时间做一次平滑
		// 例如统计周期本来设置为1分钟，但是实际上跑了2分钟，2分钟内count=100，那么对于到1分钟应该是count=50
		metric.Count = uint32(float64(data.count) * diffRate)

		if data.totalTime > 0 {
			// 有些并不需要统计平均耗时
			metric.AvgTime = data.totalTime / int64(data.count)
		}
	}

	return metric
}

// TODO 跨级别改变：向上允许跳级，但是向下不允许跳级
// 跳级的时候，例如从L1跳级到L3，则L2,L3的action都需要执行
// 如果级别发生了变化，则需要执行相应的action
func parseLevel(oldLevel, newLevel DegradeLevel, conf *TDegradeConf) DegradeLevel {
	if action := conf.actions[newLevel]; action != nil {
		// 执行相应的动作
		conf.actions[newLevel]()
	}

	return newLevel
}
