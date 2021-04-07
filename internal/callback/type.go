package callback

type Callback struct {
	Command int      `json:"C"`
	Args    []string `json:"A"`
	Index   []int    `json:"I"`
	Flags   []int    `json:"F"`
}

const (
	CommandExtractChoose = iota
	CommandExtractDone

	FlagWildcard
)

// HasArg checks whether the callback arguments include the given one
func (c Callback) HasArg(arg string) bool {
	for _, v := range c.Args {
		if v == arg {
			return true
		}
	}
	return false
}

// HasArg checks whether the callback flags include the given one
func (c Callback) HasFlag(flag int) bool {
	for _, v := range c.Flags {
		if v == flag {
			return true
		}
	}
	return false
}

// HasArg checks whether the callback indexes include the given one
func (c Callback) HasIndex(index int) bool {
	for _, v := range c.Index {
		if v == index {
			return true
		}
	}
	return false
}
