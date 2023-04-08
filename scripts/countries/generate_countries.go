package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func readJson(obj any) error {

	file, err := ioutil.ReadFile("./scripts/countries/countries.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(file, obj); err != nil {
		return err
	}

	return nil
}

var countries = []*struct {
	Name string `json:"name"`
}{}

func processCountries(list map[string]int) {

	for search, _ := range list {

		f := false
		for i, c := range countries {
			if strings.ToLower(c.Name) == strings.ToLower(search) {
				list[search] = i
				f = true
				break
			}
		}
		if !f {
			panic("country not found: " + search)
		}
	}
}

func printCountries(name string, list map[string]int) {
	fmt.Println("var " + name + " = map[uint64]bool{")
	for a, b := range list {
		fmt.Println(strconv.Itoa(b) + ":true,    //" + strings.ToLower(a))
	}
	fmt.Println("}")
}

func main() {

	if err := readJson(&countries); err != nil {
		panic(err)
	}

	europeanUnionCountries := map[string]int{
		"austria":        0,
		"belgium":        0,
		"bulgaria":       0,
		"croatia":        0,
		"cyprus":         0,
		"czech republic": 0,
		"denmark":        0,
		"estonia":        0,
		"finland":        0,
		"france":         0,
		"germany":        0,
		"greece":         0,
		"hungary":        0,
		"ireland":        0,
		"italy":          0,
		"latvia":         0,
		"lithuania":      0,
		"luxembourg":     0,
		"malta":          0,
		"netherlands":    0,
		"poland":         0,
		"portugal":       0,
		"romania":        0,
		"slovakia":       0,
		"slovenia":       0,
		"spain":          0,
		"sweden":         0,
	}

	europeanCountries := map[string]int{
		"åland islands":          0,
		"albania":                0,
		"andorra":                0,
		"austria":                0,
		"belarus":                0,
		"belgium":                0,
		"bosnia and herzegovina": 0,
		"bulgaria":               0,
		"croatia":                0,
		"czech republic":         0,
		"denmark":                0,
		"estonia":                0,
		"finland":                0,
		"france":                 0,
		"germany":                0,
		"greece":                 0,
		"holy see":               0,
		"hungary":                0,
		"ireland":                0,
		"italy":                  0,
		"latvia":                 0,
		"lithuania":              0,
		"luxembourg":             0,
		"moldova, republic of":   0,
		"malta":                  0,
		"monaco":                 0,
		"netherlands":            0,
		"montenegro":             0,
		"north macedonia":        0,
		"norway":                 0,
		"poland":                 0,
		"romania":                0,
		"russia":                 0,
		"san marino":             0,
		"serbia":                 0,
		"slovakia":               0,
		"slovenia":               0,
		"spain":                  0,
		"sweden":                 0,
		"switzerland":            0,
		"united kingdom":         0,
	}

	middleEastCountries := map[string]int{
		"bahrain":              0,
		"cyprus":               0,
		"egypt":                0,
		"iran":                 0,
		"iraq":                 0,
		"israel":               0,
		"jordan":               0,
		"kuwait":               0,
		"lebanon":              0,
		"oman":                 0,
		"palestine":            0,
		"qatar":                0,
		"saudi arabia":         0,
		"syria":                0,
		"turkey":               0,
		"united arab emirates": 0,
		"yemen":                0,
	}

	var africanCountries = map[string]int{
		"algeria":                    0, //north africa
		"libya":                      0,
		"morocco":                    0,
		"tunisia":                    0,
		"western sahara":             0,
		"burundi":                    0, //east africa
		"comoros":                    0,
		"djibouti":                   0,
		"eritrea":                    0,
		"ethiopia":                   0,
		"kenya":                      0,
		"madagascar":                 0,
		"malawi":                     0,
		"mauritius":                  0,
		"mozambique":                 0,
		"reunion":                    0,
		"rwanda":                     0,
		"seychelles":                 0,
		"somalia":                    0,
		"sudan":                      0,
		"tanzania":                   0,
		"uganda":                     0,
		"zambia":                     0,
		"zimbabwe":                   0,
		"angola":                     0, //central africa
		"cameroon":                   0,
		"central african republic":   0,
		"chad":                       0,
		"congo":                      0,
		"congo, democratic republic": 0,
		"equatorial guinea":          0,
		"gabon":                      0,
		"sao tome and principe":      0,
		"botswana":                   0, //southern africa
		"eswatini":                   0,
		"lesotho":                    0,
		"namibia":                    0,
		"south africa":               0,
		"benin":                      0, //west africa
		"burkina faso":               0,
		"cape verde":                 0,
		"gambia":                     0,
		"ghana":                      0,
		"guinea":                     0,
		"guinea-bissau":              0,
		"Cote D\"Ivoire":             0,
		"liberia":                    0,
		"mali":                       0,
		"mauritania":                 0,
		"niger":                      0,
		"nigeria":                    0,
		"saint helena":               0,
		"senegal":                    0,
		"sierra leone":               0,
		"togo":                       0,
	}

	asianCountries := map[string]int{
		"afghanistan":  0,
		"armenia":      0,
		"azerbaijan":   0,
		"bangladesh":   0,
		"bhutan":       0,
		"brunei":       0,
		"cambodia":     0,
		"china":        0,
		"georgia":      0,
		"timor-leste":  0,
		"indonesia":    0,
		"japan":        0,
		"kazakhstan":   0,
		"kuwait":       0,
		"kyrgyzstan":   0,
		"laos":         0,
		"malaysia":     0,
		"maldives":     0,
		"mongolia":     0,
		"myanmar":      0,
		"north korea":  0,
		"pakistan":     0,
		"philippines":  0,
		"saudi arabia": 0,
		"singapore":    0,
		"south korea":  0,
		"sri lanka":    0,
		"taiwan":       0,
		"tajikistan":   0,
		"thailand":     0,
		"turkmenistan": 0,
		"vietnam":      0,
	}

	oceaniaCountries := map[string]int{
		"american samoa":           0,
		"australia":                0,
		"cook islands":             0,
		"micronesia":               0,
		"fiji":                     0,
		"french polynesia":         0,
		"guam":                     0,
		"kiribati":                 0,
		"marshall islands":         0,
		"nauru":                    0,
		"new caledonia":            0,
		"new zealand":              0,
		"niue":                     0,
		"norfolk island":           0,
		"northern mariana islands": 0,
		"palau":                    0,
		"papua new guinea":         0,
		"pitcairn":                 0,
		"samoa":                    0,
		"solomon islands":          0,
		"tokelau":                  0,
		"tonga":                    0,
		"tuvalu":                   0,
		"vanuatu":                  0,
		"wallis and futuna":        0,
	}

	northAmericaCountries := map[string]int{
		"anguilla":                         0,
		"antigua and barbuda":              0,
		"aruba":                            0,
		"bahamas":                          0,
		"barbados":                         0,
		"belize":                           0,
		"bermuda":                          0,
		"virgin islands, british":          0,
		"canada":                           0,
		"cayman islands":                   0,
		"costa rica":                       0,
		"curaçao":                          0,
		"dominica":                         0,
		"dominican republic":               0,
		"el salvador":                      0,
		"greenland":                        0,
		"grenada":                          0,
		"guadeloupe":                       0,
		"guatemala":                        0,
		"haiti":                            0,
		"honduras":                         0,
		"jamaica":                          0,
		"martinique":                       0,
		"mexico":                           0,
		"montserrat":                       0,
		"nicaragua":                        0,
		"panama":                           0,
		"puerto rico":                      0,
		"saint kitts and nevis":            0,
		"saint lucia":                      0,
		"saint pierre and miquelon":        0,
		"saint vincent and the grenadines": 0,
		"trinidad and tobago":              0,
		"turks and caicos islands":         0,
		"united states":                    0,
		"virgin islands, u.s.":             0,
	}

	southAmericaCountries := map[string]int{
		"argentina":                   0,
		"bolivia":                     0,
		"brazil":                      0,
		"chile":                       0,
		"colombia":                    0,
		"ecuador":                     0,
		"falkland islands (malvinas)": 0,
		"french guiana":               0,
		"guyana":                      0,
		"paraguay":                    0,
		"peru":                        0,
		"south georgia":               0,
		"suriname":                    0,
		"uruguay":                     0,
		"venezuela":                   0,
	}

	processCountries(europeanUnionCountries)
	processCountries(europeanCountries)
	processCountries(middleEastCountries)
	processCountries(africanCountries)
	processCountries(asianCountries)
	processCountries(oceaniaCountries)
	processCountries(northAmericaCountries)
	processCountries(southAmericaCountries)
	printCountries("EuropeanUnionCountries", europeanUnionCountries)
	printCountries("EuropeanCountries", europeanCountries)
	printCountries("MiddleEastCountries", middleEastCountries)
	printCountries("AfricanCountries", africanCountries)
	printCountries("OceaniaCountries", oceaniaCountries)
	printCountries("NorthAmericaCountries", northAmericaCountries)
	printCountries("SouthAmericaCountries", southAmericaCountries)

}
