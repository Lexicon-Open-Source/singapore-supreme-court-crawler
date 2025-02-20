// This file is generated by "./lib/proto/generate"

package proto

/*

HeapProfiler

*/

// HeapProfilerHeapSnapshotObjectID Heap snapshot object id.
type HeapProfilerHeapSnapshotObjectID string

// HeapProfilerSamplingHeapProfileNode Sampling Heap Profile node. Holds callsite information, allocation statistics and child nodes.
type HeapProfilerSamplingHeapProfileNode struct {
	// CallFrame Function location.
	CallFrame *RuntimeCallFrame `json:"callFrame"`

	// SelfSize Allocations size in bytes for the node excluding children.
	SelfSize float64 `json:"selfSize"`

	// ID Node id. Ids are unique across all profiles collected between startSampling and stopSampling.
	ID int `json:"id"`

	// Children Child nodes.
	Children []*HeapProfilerSamplingHeapProfileNode `json:"children"`
}

// HeapProfilerSamplingHeapProfileSample A single sample from a sampling profile.
type HeapProfilerSamplingHeapProfileSample struct {
	// Size Allocation size in bytes attributed to the sample.
	Size float64 `json:"size"`

	// NodeID Id of the corresponding profile tree node.
	NodeID int `json:"nodeId"`

	// Ordinal Time-ordered sample ordinal number. It is unique across all profiles retrieved
	// between startSampling and stopSampling.
	Ordinal float64 `json:"ordinal"`
}

// HeapProfilerSamplingHeapProfile Sampling profile.
type HeapProfilerSamplingHeapProfile struct {
	// Head ...
	Head *HeapProfilerSamplingHeapProfileNode `json:"head"`

	// Samples ...
	Samples []*HeapProfilerSamplingHeapProfileSample `json:"samples"`
}

// HeapProfilerAddInspectedHeapObject Enables console to refer to the node with given id via $x (see Command Line API for more details
// $x functions).
type HeapProfilerAddInspectedHeapObject struct {
	// HeapObjectID Heap snapshot object id to be accessible by means of $x command line API.
	HeapObjectID HeapProfilerHeapSnapshotObjectID `json:"heapObjectId"`
}

// ProtoReq name.
func (m HeapProfilerAddInspectedHeapObject) ProtoReq() string {
	return "HeapProfiler.addInspectedHeapObject"
}

// Call sends the request.
func (m HeapProfilerAddInspectedHeapObject) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerCollectGarbage ...
type HeapProfilerCollectGarbage struct{}

// ProtoReq name.
func (m HeapProfilerCollectGarbage) ProtoReq() string { return "HeapProfiler.collectGarbage" }

// Call sends the request.
func (m HeapProfilerCollectGarbage) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerDisable ...
type HeapProfilerDisable struct{}

// ProtoReq name.
func (m HeapProfilerDisable) ProtoReq() string { return "HeapProfiler.disable" }

// Call sends the request.
func (m HeapProfilerDisable) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerEnable ...
type HeapProfilerEnable struct{}

// ProtoReq name.
func (m HeapProfilerEnable) ProtoReq() string { return "HeapProfiler.enable" }

// Call sends the request.
func (m HeapProfilerEnable) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerGetHeapObjectID ...
type HeapProfilerGetHeapObjectID struct {
	// ObjectID Identifier of the object to get heap object id for.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
}

// ProtoReq name.
func (m HeapProfilerGetHeapObjectID) ProtoReq() string { return "HeapProfiler.getHeapObjectId" }

// Call the request.
func (m HeapProfilerGetHeapObjectID) Call(c Client) (*HeapProfilerGetHeapObjectIDResult, error) {
	var res HeapProfilerGetHeapObjectIDResult
	return &res, call(m.ProtoReq(), m, &res, c)
}

// HeapProfilerGetHeapObjectIDResult ...
type HeapProfilerGetHeapObjectIDResult struct {
	// HeapSnapshotObjectID Id of the heap snapshot object corresponding to the passed remote object id.
	HeapSnapshotObjectID HeapProfilerHeapSnapshotObjectID `json:"heapSnapshotObjectId"`
}

// HeapProfilerGetObjectByHeapObjectID ...
type HeapProfilerGetObjectByHeapObjectID struct {
	// ObjectID ...
	ObjectID HeapProfilerHeapSnapshotObjectID `json:"objectId"`

	// ObjectGroup (optional) Symbolic group name that can be used to release multiple objects.
	ObjectGroup string `json:"objectGroup,omitempty"`
}

// ProtoReq name.
func (m HeapProfilerGetObjectByHeapObjectID) ProtoReq() string {
	return "HeapProfiler.getObjectByHeapObjectId"
}

// Call the request.
func (m HeapProfilerGetObjectByHeapObjectID) Call(c Client) (*HeapProfilerGetObjectByHeapObjectIDResult, error) {
	var res HeapProfilerGetObjectByHeapObjectIDResult
	return &res, call(m.ProtoReq(), m, &res, c)
}

// HeapProfilerGetObjectByHeapObjectIDResult ...
type HeapProfilerGetObjectByHeapObjectIDResult struct {
	// Result Evaluation result.
	Result *RuntimeRemoteObject `json:"result"`
}

// HeapProfilerGetSamplingProfile ...
type HeapProfilerGetSamplingProfile struct{}

// ProtoReq name.
func (m HeapProfilerGetSamplingProfile) ProtoReq() string { return "HeapProfiler.getSamplingProfile" }

// Call the request.
func (m HeapProfilerGetSamplingProfile) Call(c Client) (*HeapProfilerGetSamplingProfileResult, error) {
	var res HeapProfilerGetSamplingProfileResult
	return &res, call(m.ProtoReq(), m, &res, c)
}

// HeapProfilerGetSamplingProfileResult ...
type HeapProfilerGetSamplingProfileResult struct {
	// Profile Return the sampling profile being collected.
	Profile *HeapProfilerSamplingHeapProfile `json:"profile"`
}

// HeapProfilerStartSampling ...
type HeapProfilerStartSampling struct {
	// SamplingInterval (optional) Average sample interval in bytes. Poisson distribution is used for the intervals. The
	// default value is 32768 bytes.
	SamplingInterval *float64 `json:"samplingInterval,omitempty"`

	// IncludeObjectsCollectedByMajorGC (optional) By default, the sampling heap profiler reports only objects which are
	// still alive when the profile is returned via getSamplingProfile or
	// stopSampling, which is useful for determining what functions contribute
	// the most to steady-state memory usage. This flag instructs the sampling
	// heap profiler to also include information about objects discarded by
	// major GC, which will show which functions cause large temporary memory
	// usage or long GC pauses.
	IncludeObjectsCollectedByMajorGC bool `json:"includeObjectsCollectedByMajorGC,omitempty"`

	// IncludeObjectsCollectedByMinorGC (optional) By default, the sampling heap profiler reports only objects which are
	// still alive when the profile is returned via getSamplingProfile or
	// stopSampling, which is useful for determining what functions contribute
	// the most to steady-state memory usage. This flag instructs the sampling
	// heap profiler to also include information about objects discarded by
	// minor GC, which is useful when tuning a latency-sensitive application
	// for minimal GC activity.
	IncludeObjectsCollectedByMinorGC bool `json:"includeObjectsCollectedByMinorGC,omitempty"`
}

// ProtoReq name.
func (m HeapProfilerStartSampling) ProtoReq() string { return "HeapProfiler.startSampling" }

// Call sends the request.
func (m HeapProfilerStartSampling) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerStartTrackingHeapObjects ...
type HeapProfilerStartTrackingHeapObjects struct {
	// TrackAllocations (optional) ...
	TrackAllocations bool `json:"trackAllocations,omitempty"`
}

// ProtoReq name.
func (m HeapProfilerStartTrackingHeapObjects) ProtoReq() string {
	return "HeapProfiler.startTrackingHeapObjects"
}

// Call sends the request.
func (m HeapProfilerStartTrackingHeapObjects) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerStopSampling ...
type HeapProfilerStopSampling struct{}

// ProtoReq name.
func (m HeapProfilerStopSampling) ProtoReq() string { return "HeapProfiler.stopSampling" }

// Call the request.
func (m HeapProfilerStopSampling) Call(c Client) (*HeapProfilerStopSamplingResult, error) {
	var res HeapProfilerStopSamplingResult
	return &res, call(m.ProtoReq(), m, &res, c)
}

// HeapProfilerStopSamplingResult ...
type HeapProfilerStopSamplingResult struct {
	// Profile Recorded sampling heap profile.
	Profile *HeapProfilerSamplingHeapProfile `json:"profile"`
}

// HeapProfilerStopTrackingHeapObjects ...
type HeapProfilerStopTrackingHeapObjects struct {
	// ReportProgress (optional) If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken
	// when the tracking is stopped.
	ReportProgress bool `json:"reportProgress,omitempty"`

	// TreatGlobalObjectsAsRoots (deprecated) (optional) Deprecated in favor of `exposeInternals`.
	TreatGlobalObjectsAsRoots bool `json:"treatGlobalObjectsAsRoots,omitempty"`

	// CaptureNumericValue (optional) If true, numerical values are included in the snapshot
	CaptureNumericValue bool `json:"captureNumericValue,omitempty"`

	// ExposeInternals (experimental) (optional) If true, exposes internals of the snapshot.
	ExposeInternals bool `json:"exposeInternals,omitempty"`
}

// ProtoReq name.
func (m HeapProfilerStopTrackingHeapObjects) ProtoReq() string {
	return "HeapProfiler.stopTrackingHeapObjects"
}

// Call sends the request.
func (m HeapProfilerStopTrackingHeapObjects) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerTakeHeapSnapshot ...
type HeapProfilerTakeHeapSnapshot struct {
	// ReportProgress (optional) If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken.
	ReportProgress bool `json:"reportProgress,omitempty"`

	// TreatGlobalObjectsAsRoots (deprecated) (optional) If true, a raw snapshot without artificial roots will be generated.
	// Deprecated in favor of `exposeInternals`.
	TreatGlobalObjectsAsRoots bool `json:"treatGlobalObjectsAsRoots,omitempty"`

	// CaptureNumericValue (optional) If true, numerical values are included in the snapshot
	CaptureNumericValue bool `json:"captureNumericValue,omitempty"`

	// ExposeInternals (experimental) (optional) If true, exposes internals of the snapshot.
	ExposeInternals bool `json:"exposeInternals,omitempty"`
}

// ProtoReq name.
func (m HeapProfilerTakeHeapSnapshot) ProtoReq() string { return "HeapProfiler.takeHeapSnapshot" }

// Call sends the request.
func (m HeapProfilerTakeHeapSnapshot) Call(c Client) error {
	return call(m.ProtoReq(), m, nil, c)
}

// HeapProfilerAddHeapSnapshotChunk ...
type HeapProfilerAddHeapSnapshotChunk struct {
	// Chunk ...
	Chunk string `json:"chunk"`
}

// ProtoEvent name.
func (evt HeapProfilerAddHeapSnapshotChunk) ProtoEvent() string {
	return "HeapProfiler.addHeapSnapshotChunk"
}

// HeapProfilerHeapStatsUpdate If heap objects tracking has been started then backend may send update for one or more fragments.
type HeapProfilerHeapStatsUpdate struct {
	// StatsUpdate An array of triplets. Each triplet describes a fragment. The first integer is the fragment
	// index, the second integer is a total count of objects for the fragment, the third integer is
	// a total size of the objects for the fragment.
	StatsUpdate []int `json:"statsUpdate"`
}

// ProtoEvent name.
func (evt HeapProfilerHeapStatsUpdate) ProtoEvent() string {
	return "HeapProfiler.heapStatsUpdate"
}

// HeapProfilerLastSeenObjectID If heap objects tracking has been started then backend regularly sends a current value for last
// seen object id and corresponding timestamp. If the were changes in the heap since last event
// then one or more heapStatsUpdate events will be sent before a new lastSeenObjectId event.
type HeapProfilerLastSeenObjectID struct {
	// LastSeenObjectID ...
	LastSeenObjectID int `json:"lastSeenObjectId"`

	// Timestamp ...
	Timestamp float64 `json:"timestamp"`
}

// ProtoEvent name.
func (evt HeapProfilerLastSeenObjectID) ProtoEvent() string {
	return "HeapProfiler.lastSeenObjectId"
}

// HeapProfilerReportHeapSnapshotProgress ...
type HeapProfilerReportHeapSnapshotProgress struct {
	// Done ...
	Done int `json:"done"`

	// Total ...
	Total int `json:"total"`

	// Finished (optional) ...
	Finished bool `json:"finished,omitempty"`
}

// ProtoEvent name.
func (evt HeapProfilerReportHeapSnapshotProgress) ProtoEvent() string {
	return "HeapProfiler.reportHeapSnapshotProgress"
}

// HeapProfilerResetProfiles ...
type HeapProfilerResetProfiles struct{}

// ProtoEvent name.
func (evt HeapProfilerResetProfiles) ProtoEvent() string {
	return "HeapProfiler.resetProfiles"
}
