package domain

// Verdict Вердикт технического осмотра
type Verdict struct {
	Approved bool
	// Reasons - за что отказано
	Reasons []string
}

// Approve — Принят
func Approve() Verdict { return Verdict{Approved: true} }

// Reject — Отказ с перечнем причин
func Reject(reasons ...string) Verdict {
	return Verdict{Approved: false, Reasons: reasons}
}
