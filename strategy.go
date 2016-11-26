// 策略

package crondStrategy

import (
	"sync"
)

// 级别的类型
type TLevel uint8

// 5个级别的类型
const (
	levelNotInit TLevel = iota // 未初始化的级别，只是内部使用
	Level1
	Level2
	Level3
	Level4
	Level5
)

// 只有两个级别的策略常量，对应true and false
const (
	StatusYes = Level1
	StatusNo  = Level2
)

// 级别的总数
const LevelCount = 5

// 策略配置结构
type TStrategy struct {
	// 标识该策略是否已经关闭
	isClose bool

	// 最后的level
	lastLevel TLevel

	// 降级策略的判断函数
	// 参数为各统计指标
	Check func(*TEventsMetric) TLevel

	// 判断之后执行的函数
	Action func(TLevel)
}

type tStrategyList struct {
	sync.Mutex
	strategy []*TStrategy
}

var (
	strategyList *tStrategyList
)

func init() {
	strategyList = &tStrategyList{}
}

// AddConf 增加执行策略配置
func NewStrategy(check func(*TEventsMetric) TLevel, action func(TLevel)) (strategy *TStrategy) {
	strategy = &TStrategy{
		Check:  check,
		Action: action,
	}

	strategyList.Lock()
	strategyList.strategy = append(strategyList.strategy, strategy)
	strategyList.Unlock()

	return strategy
}

// 停止策略
func (s *TStrategy) Stop() {
	strategyList.Lock()
	s.isClose = true
	s.lastLevel = levelNotInit
	strategyList.Unlock()
}

// 启动策略
func (s *TStrategy) Start() {
	strategyList.Lock()
	s.isClose = false
	strategyList.Unlock()
}
