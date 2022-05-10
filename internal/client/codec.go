package client

type ScaleEnum map[string][]byte

func (s *ScaleEnum) Value() string {
	for k := range *s {
		return k
	}
	return ""
}
