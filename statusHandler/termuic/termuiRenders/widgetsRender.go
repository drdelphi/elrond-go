package termuiRenders

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const statusSynching = "currently synching"
const statusSynchronized = "synchronized"

//WidgetsRender will define termui widgets that need to display a termui console
type WidgetsRender struct {
	container    *DrawableContainer
	lLog         *widgets.List
	instanceInfo *widgets.Table
	chainInfo    *widgets.Table

	cpuLoad     *widgets.Gauge
	memoryLoad  *widgets.Gauge
	networkRecv *widgets.Gauge
	networkSent *widgets.Gauge

	termuiConsoleMetrics *sync.Map
}

//NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender(metricData *sync.Map, grid *DrawableContainer) *WidgetsRender {
	self := &WidgetsRender{
		termuiConsoleMetrics: metricData,
		container:            grid,
	}
	self.initWidgets()
	self.setGrid()

	return self
}

func (wr *WidgetsRender) initWidgets() {
	wr.instanceInfo = widgets.NewTable()
	wr.instanceInfo.Rows = [][]string{{""}}

	wr.chainInfo = widgets.NewTable()
	wr.chainInfo.Rows = [][]string{{"", "", "", ""}}

	wr.cpuLoad = widgets.NewGauge()
	wr.memoryLoad = widgets.NewGauge()
	wr.networkRecv = widgets.NewGauge()
	wr.networkSent = widgets.NewGauge()

	wr.lLog = widgets.NewList()
}

func (wr *WidgetsRender) setGrid() {
	gridLeft := ui.NewGrid()

	gridLeft.Set(
		ui.NewRow(10.0/22, wr.instanceInfo),
		ui.NewRow(12.0/22, wr.chainInfo))

	gridRight := ui.NewGrid()
	gridRight.Set(ui.NewRow(1.0/4, wr.cpuLoad),
		ui.NewRow(1.0/4, wr.memoryLoad),
		ui.NewRow(1.0/4, wr.networkRecv),
		ui.NewRow(1.0/4, wr.networkSent))

	gridBottom := ui.NewGrid()
	gridBottom.Set(ui.NewRow(1.0, wr.lLog))

	wr.container.SetTopLeft(gridLeft)
	wr.container.SetTopRight(gridRight)
	wr.container.SetBottom(gridBottom)
}

//RefreshData method is used to prepare data that are displayed on container
func (wr *WidgetsRender) RefreshData(logLines []string) {
	wr.prepareInstanceInfo()
	wr.prepareChainInfo()
	wr.prepareListWithLogsForDisplay(logLines)
	wr.prepareLoads()
}

func (wr *WidgetsRender) prepareInstanceInfo() {
	//8 rows and one column
	numRows := 8
	rows := make([][]string, numRows)

	appVersion := wr.getFromCacheAsString(core.MetricAppVersion)
	rows[0] = []string{fmt.Sprintf("App version: %s", appVersion)}
	if strings.Contains(appVersion, core.UnVersionedAppString) {
		wr.instanceInfo.RowStyles[0] = ui.NewStyle(ui.ColorRed)
	} else {
		wr.instanceInfo.RowStyles[0] = ui.NewStyle(ui.ColorGreen)
	}

	pkTxSign := wr.getFromCacheAsString(core.MetricPublicKeyTxSign)
	rows[1] = []string{fmt.Sprintf("Public key TxSign: %s", pkTxSign)}

	pkBlockSign := wr.getFromCacheAsString(core.MetricPublicKeyBlockSign)
	rows[2] = []string{fmt.Sprintf("Public key BlockSign: %s", pkBlockSign)}

	shardId := wr.getFromCacheAsUint64(core.MetricShardId)
	shardIdStr := ""
	if shardId == uint64(sharding.MetachainShardId) {
		shardIdStr = "meta"
	} else {
		shardIdStr = fmt.Sprintf("%d", shardId)
	}
	rows[3] = []string{fmt.Sprintf("ShardID: %s", shardIdStr)}

	instanceType := wr.getFromCacheAsString(core.MetricNodeType)
	rows[4] = []string{fmt.Sprintf("Instance type: %s", instanceType)}

	countConsensus := wr.getFromCacheAsUint64(core.MetricCountConsensus)
	rows[5] = []string{fmt.Sprintf("Consensus group participation count: %d", countConsensus)}

	countLeader := wr.getFromCacheAsUint64(core.MetricCountLeader)
	rows[6] = []string{fmt.Sprintf("Elected consensus leader count: %d", countLeader)}

	countAcceptedBlocks := wr.getFromCacheAsUint64(core.MetricCountAcceptedBlocks)
	rows[7] = []string{fmt.Sprintf("Consensus proposed & accepted blocks: %v", countAcceptedBlocks)}

	wr.instanceInfo.Title = "Instance info"
	wr.instanceInfo.RowSeparator = false
	wr.instanceInfo.Rows = rows
}

func (wr *WidgetsRender) prepareChainInfo() {
	//10 rows and one column
	numRows := 10
	rows := make([][]string, numRows)

	syncStatus := wr.getFromCacheAsUint64(core.MetricIsSyncing)
	syncingStr := fmt.Sprintf("undefined %d", syncStatus)
	switch syncStatus {
	case 1:
		syncingStr = statusSynching
	case 0:
		syncingStr = statusSynchronized
	}
	rows[0] = []string{fmt.Sprintf("Status: %s", syncingStr)}
	if strings.Contains(syncingStr, statusSynchronized) {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorGreen)
	} else {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorBlue)
	}

	memTxPoolSize := wr.getFromCacheAsUint64(core.MetricTxPoolLoad)
	rows[1] = []string{fmt.Sprintf("Number of transactions in pool: %d", memTxPoolSize)}

	rows[2] = make([]string, 0)

	nonce := wr.getFromCacheAsUint64(core.MetricNonce)
	probableHighestNonce := wr.getFromCacheAsUint64(core.MetricProbableHighestNonce)
	rows[3] = []string{fmt.Sprintf("Current synchronized block nonce: %v / %v",
		nonce, probableHighestNonce)}

	synchronizedRound := wr.getFromCacheAsUint64(core.MetricSynchronizedRound)
	currentRound := wr.getFromCacheAsUint64(core.MetricCurrentRound)
	rows[4] = []string{fmt.Sprintf("Current consensus round: %v / %v",
		synchronizedRound, currentRound)}

	consensusRoundTime := wr.getFromCacheAsUint64(core.MetricRoundTime)
	rows[5] = []string{fmt.Sprintf("Consensus round time: %ds", consensusRoundTime)}

	rows[6] = make([]string, 0)

	numLiveValidators := wr.getFromCacheAsUint64(core.MetricLiveValidatorNodes)
	rows[7] = []string{fmt.Sprintf("Live validator nodes: %v", numLiveValidators)}

	numConnectedNodes := wr.getFromCacheAsUint64(core.MetricConnectedNodes)
	rows[8] = []string{fmt.Sprintf("Network connected nodes: %v", numConnectedNodes)}

	numConnectedPeers := wr.getFromCacheAsUint64(core.MetricNumConnectedPeers)
	rows[9] = []string{fmt.Sprintf("This node is connected to %v peers", numConnectedPeers)}

	wr.chainInfo.Title = "Chain info"
	wr.chainInfo.RowSeparator = false
	wr.chainInfo.Rows = rows
}

func (wr *WidgetsRender) prepareListWithLogsForDisplay(logData []string) {
	wr.lLog.Title = "Log info"
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorWhite)
	wr.lLog.Rows = wr.prepareLogLines(logData, wr.lLog.Size().Y)
	wr.lLog.WrapText = true
}

//TODO duplicate code next pull request refactor
func (wr *WidgetsRender) prepareLogLines(logData []string, size int) []string {
	logDataLen := len(logData)
	if logDataLen > size {
		return logData[logDataLen-size : logDataLen]
	}

	return logData
}

func (wr *WidgetsRender) prepareLoads() {
	cpuLoadPercent := wr.getFromCacheAsUint64(core.MetricCpuLoadPercent)
	wr.cpuLoad.Title = "CPU Load"
	wr.cpuLoad.Percent = int(cpuLoadPercent)

	memLoadPercent := wr.getFromCacheAsUint64(core.MetricMemLoadPercent)
	memTotalMemoryBytes := wr.getFromCacheAsUint64(core.MetricTotalMem)
	wr.memoryLoad.Title = "Memory load"
	wr.memoryLoad.Percent = int(memLoadPercent)
	wr.memoryLoad.Label = fmt.Sprintf("%d%% / total: %s", memLoadPercent, core.ConvertBytes(memTotalMemoryBytes))

	recvLoad := wr.getFromCacheAsUint64(core.MetricNetworkRecvPercent)
	recvBps := wr.getFromCacheAsUint64(core.MetricNetworkRecvBps)
	recvBpsPeak := wr.getFromCacheAsUint64(core.MetricNetworkRecvBpsPeak)
	wr.networkRecv.Title = "Network - recv:"
	wr.networkRecv.Percent = int(recvLoad)
	wr.networkRecv.Label = fmt.Sprintf("%d%% / rate: %s/s / peak rate: %s/s",
		recvLoad, core.ConvertBytes(recvBps), core.ConvertBytes(recvBpsPeak))

	sentLoad := wr.getFromCacheAsUint64(core.MetricNetworkSentPercent)
	sentBps := wr.getFromCacheAsUint64(core.MetricNetworkSentBps)
	sentBpsPeak := wr.getFromCacheAsUint64(core.MetricNetworkSentBpsPeak)
	wr.networkSent.Title = "Network - sent:"
	wr.networkSent.Percent = int(sentLoad)
	wr.networkSent.Label = fmt.Sprintf("%d%% / rate: %s/s / peak rate: %s/s",
		sentLoad, core.ConvertBytes(sentBps), core.ConvertBytes(sentBpsPeak))
}

func (wr *WidgetsRender) getFromCacheAsUint64(metric string) uint64 {
	val, ok := wr.termuiConsoleMetrics.Load(metric)
	if !ok {
		return 0
	}

	valUint64, ok := val.(uint64)
	if !ok {
		return 0
	}

	return valUint64
}

func (wr *WidgetsRender) getFromCacheAsString(metric string) string {
	val, ok := wr.termuiConsoleMetrics.Load(metric)
	if !ok {
		return "[invalid key]"
	}

	valStr, ok := val.(string)
	if !ok {
		return "[not a string]"
	}

	return valStr
}

func (wr *WidgetsRender) getNetworkRecvStats() {
}
