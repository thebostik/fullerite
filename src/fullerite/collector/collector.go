package collector

import (
	"fullerite/config"
	"fullerite/metric"

	"regexp"
	"strings"

	l "github.com/Sirupsen/logrus"
)

const (
	// DefaultCollectionInterval the interval to collect on unless overridden by a collectors config
	DefaultCollectionInterval = 10
)

var defaultLog = l.WithFields(l.Fields{"app": "fullerite", "pkg": "collector"})

// Collector defines the interface of a generic collector.
type Collector interface {
	Collect()
	Configure(map[string]interface{})

	// taken care of by the base class
	Name() string
	Channel() chan metric.Metric
	Interval() int
	SetInterval(int)
	CollectorType() string
	SetCollectorType(string)
	CanonicalName() string
	SetCanonicalName(string)
	Prefix() string
	SetPrefix(string)
	Blacklist() []string
	SetBlacklist([]string)
	DimensionsBlacklist() map[string]string
	SetDimensionsBlacklist(map[string]string)
	ContainsBlacklistedDimension(map[string]string) bool
}

var collectorConstructs map[string]func(chan metric.Metric, int, *l.Entry) Collector

// RegisterCollector composes a map of collector names -> factor functions
func RegisterCollector(name string, f func(chan metric.Metric, int, *l.Entry) Collector) {
	if collectorConstructs == nil {
		collectorConstructs = make(map[string]func(chan metric.Metric, int, *l.Entry) Collector)
	}
	collectorConstructs[name] = f
}

// New creates a new Collector based on the requested collector name.
func New(name string) Collector {
	var collector Collector

	channel := make(chan metric.Metric)
	collectorLog := defaultLog.WithFields(l.Fields{"collector": name})
	// This allows for initiating multiple collectors of the same type
	// but with a different canonical name so they can receive different
	// configs
	realName := strings.Split(name, " ")[0]

	if f, exists := collectorConstructs[realName]; exists {
		collector = f(channel, DefaultCollectionInterval, collectorLog)
	} else {
		defaultLog.Error("Cannot create collector: ", realName)
		return nil
	}

	if collector.CollectorType() == "" {
		collector.SetCollectorType("collector")
	}
	collector.SetCanonicalName(name)
	return collector
}

type baseCollector struct {
	// fulfill most of the rote parts of the collector interface
	channel             chan metric.Metric
	name                string
	interval            int
	collectorType       string
	canonicalName       string
	prefix              string
	blacklist           []string
	dimensionsBlacklist map[string]string

	// intentionally exported
	log *l.Entry
}

func (col *baseCollector) configureCommonParams(configMap map[string]interface{}) {
	if interval, exists := configMap["interval"]; exists {
		col.interval = config.GetAsInt(interval, DefaultCollectionInterval)
	}

	if prefix, exists := configMap["prefix"]; exists {
		if str, ok := prefix.(string); ok {
			col.prefix = str
		}
	}

	if asInterface, exists := configMap["metrics_blacklist"]; exists {
		col.blacklist = config.GetAsSlice(asInterface)
	}

	if asInterface, exists := configMap["dimensions_blacklist"]; exists {
		col.dimensionsBlacklist = config.GetAsMap(asInterface)
	}
}

// SetInterval : set the interval to collect on
func (col *baseCollector) SetInterval(interval int) {
	col.interval = interval
}

// SetPrefix : set the optional prefix for the collector
func (col *baseCollector) SetPrefix(prefix string) {
	col.prefix = prefix
}

// SetCollectorType : collector type
func (col *baseCollector) SetCollectorType(collectorType string) {
	col.collectorType = collectorType
}

// SetCanonicalName : collector canonical name
func (col *baseCollector) SetCanonicalName(name string) {
	col.canonicalName = name
}

// SetBlacklist : set collector optional metrics blacklist
func (col *baseCollector) SetBlacklist(blacklist []string) {
	col.blacklist = blacklist
}

// SetDimensionsBlacklist : set collector optional dimensions blacklist
func (col *baseCollector) SetDimensionsBlacklist(blacklist map[string]string) {
	col.dimensionsBlacklist = blacklist
}

// CanonicalName : collector canonical name
func (col *baseCollector) CanonicalName() string {
	return col.canonicalName
}

// CollectorType : collector type
func (col *baseCollector) CollectorType() string {
	return col.collectorType
}

// Prefix : return optional prefix for all metrics from this collector
func (col *baseCollector) Prefix() string {
	return col.prefix
}

// Channel : the channel on which the collector should send metrics
func (col baseCollector) Channel() chan metric.Metric {
	return col.channel
}

// Name : the name of the collector
func (col baseCollector) Name() string {
	return col.name
}

// Interval : the interval to collect the metrics on
func (col baseCollector) Interval() int {
	return col.interval
}

// String returns the collector name in printable format.
func (col baseCollector) String() string {
	return col.Name() + "Collector"
}

// Blacklist returns the list of metrics to be blacklisted for this collector
func (col *baseCollector) Blacklist() []string {
	return col.blacklist
}

// DimensionsBlacklist returns the list of dimensions to be blacklisted for this collector
func (col *baseCollector) DimensionsBlacklist() map[string]string {
	return col.dimensionsBlacklist
}

// ContainsBlacklistedDimension returns the true if dimensions passed as argument
// contain values blacklisted by the user
func (col *baseCollector) ContainsBlacklistedDimension(dimensions map[string]string) bool {
	for k, v := range col.DimensionsBlacklist() {
		if match, err := regexp.MatchString(v, dimensions[k]); match {
			return true
		} else if err != nil {
			// Immediately return if there is any error
			break
		}
	}
	return false
}
