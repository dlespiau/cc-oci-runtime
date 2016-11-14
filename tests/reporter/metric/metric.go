package metric

type Desc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Metric interface {
	// Does the metric pass?
	Pass() bool
	// It's possible to ask the question: does this metric pass given a
	// previous reference of that metric.
	// For instance:
	//   For a boolean metric:
	//     - Pass() will always return false if the boolean metric is false
	//     - PassWithReference() will return:
	//       - true if the metric is false and the reference given is also
	//         false. If could be that the test was failing but is still
	//         failing, not a worse result
	//       - false otherwise
	//   For a coverage metric:
	//     - PassWithReference() will return true only if the new metric
	//       improves coverage over the reference.
	//   etc
	PassWithReference(ref Metric) bool
}

func FromDesc(desc Desc) Metric {

}
