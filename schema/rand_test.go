package schema

type fakeSource int64

func (s fakeSource) Int63() int64 {
	return int64(s)
}

func (s fakeSource) Seed(int64) {}
