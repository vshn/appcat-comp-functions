package runtime

import "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"

type Result interface {
	Resolve() v1alpha1.Result
}

type result v1alpha1.Result

// NewWarning results are non-fatal; the entire Composition will run to
// completion but warning events and debug logs associated with the
// composite resource will be emitted.
func NewWarning(msg string) Result {
	return result{
		Severity: v1alpha1.SeverityWarning,
		Message:  msg,
	}
}

// NewFatal results are fatal; subsequent Composition Functions may run, but
// the Composition Function pipeline run will be considered a failure and
// the first error will be returned.
func NewFatal(msg string) Result {
	return result{
		Severity: v1alpha1.SeverityFatal,
		Message:  msg,
	}
}

// NewNormal results are emitted as normal events and debug logs associated
// with the composite resource.
func NewNormal() Result {
	return result{
		Severity: v1alpha1.SeverityNormal,
		Message:  "function ran successfully",
	}
}

// Resolve returns the wrapped object v1alpha1.Result from crossplane
func (r result) Resolve() v1alpha1.Result {
	return v1alpha1.Result(r)
}
