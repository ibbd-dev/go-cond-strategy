package servicesDegrade

import (
	"sync"
	"time"
)

// 事件
type TEvent struct {
	name      string // 事件名
	beginTime int64  // 事件开始时间
}

type tEventStat struct {
	// 1分钟数值统计
	oneMinute tStatData

	// 5分钟数值统计
	fiveMinute tStatData
}

type tStatData struct {
	count     uint32 // 事件发生次数
	totalTime int64  // 时间段内总耗时。单位：纳秒
	beginTime int64  // 开始时间。单位：纳秒
}

var (
	eventsRW sync.RWMutex
	events   map[string]*tEventStat
)

var nowFunc = time.Now

// 事件必须先初始化
func InitEvent(name string) {
	eventsRW.Lock()
	if events == nil {
		events = make(map[string]*tEventStat)
	}

	// 事件初始化
	events[name] = &tEventStat{}
	events[name].oneMinute.beginTime = nowFunc().UnixNano()
	events[name].fiveMinute.beginTime = events[name].oneMinute.beginTime

	// 指标初始化
	eventsMetric.Events[name] = &TEventMetric{}

	eventsRW.Unlock()
}

// 开始一个事件
func BeginEvent(name string) *TEvent {
	return &TEvent{name: name, beginTime: nowFunc().UnixNano()}
}

// 设置事件的名字（名字变了，就是另一个事件了）
// 例如：接口访问事件，正常返回是就是访问事件，但是执行过程中发生了异常，这时可能会作为一个异常事件记录
// 事件名必须已经调用InitEvent进行初始化了
func (e *TEvent) SetName(name string) {
	e.name = name
}

// 事件结束接口
func (e *TEvent) End() {
	diff := nowFunc().UnixNano() - e.beginTime

	eventsRW.Lock()
	events[e.name].oneMinute.count++
	events[e.name].oneMinute.totalTime += diff
	eventsRW.Unlock()
}
