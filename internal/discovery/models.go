package discovery

var G1LEDModels = []string{"RSLED50", "RSLED90", "RSLED160"}
var G2LEDModels = []string{"RSLED60", "RSLED115", "RSLED170"}
var LEDModels = append(G1LEDModels, G2LEDModels...)
var DoseModels = []string{"RSDOSE2", "RSDOSE4"}
var MatModels = []string{"RSMAT", "RSMAT250", "RSMAT500", "RSMAT1200"}
var ATOModels = []string{"RSATO+"}
var RunModels = []string{"RSRUN"}
var WaveModels = []string{"RSWAVE25", "RSWAVE45"}

var AllModels []string

func init() {
	AllModels = append(AllModels, LEDModels...)
	AllModels = append(AllModels, DoseModels...)
	AllModels = append(AllModels, MatModels...)
	AllModels = append(AllModels, ATOModels...)
	AllModels = append(AllModels, RunModels...)
	AllModels = append(AllModels, WaveModels...)
}

func IsKnownModel(model string) bool {
	for _, m := range AllModels {
		if m == model {
			return true
		}
	}
	return false
}

func IsG2LED(model string) bool {
	for _, m := range G2LEDModels {
		if m == model {
			return true
		}
	}
	return false
}
