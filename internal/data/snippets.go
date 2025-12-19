package data

// Snippet represents a menu snippet with a name and content
type Snippet struct {
	Name    string
	Content string
}

// SnippetCategory represents a category of snippets
type SnippetCategory struct {
	Name     string
	Snippets []Snippet
}

// GetSnippetCategories returns all snippet categories for the menu
func GetSnippetCategories() []SnippetCategory {
	return []SnippetCategory{
		{
			Name: "Basic Math",
			Snippets: []Snippet{
				{"Arithmetic", "10 + 20 * 3 =\n"},
				{"Currency", "$1,500.00 + $250.50 =\n"},
				{"Line Reference", "100 =\n\\1 * 2 =\n"},
				{"Scientific Functions", "sin(45) + cos(30) =\nsqrt(144) =\nabs(-50) =\n"},
				{"Complex Expression", "$1,000 x 12 - 15% + $500 =\n"},
				{"Comparison", "25 > 2.5 =\n100 >= 100 =\n5 != 3 =\n"},
				{"Base Conversion", "255 in hex =\n0xFF in dec =\n25 in bin =\n0b11001 in oct =\n"},
			},
		},
		{
			Name: "Constants",
			Snippets: []Snippet{
				{"Mathematical", "pi =\ne =\nphi =\ngolden ratio =\n"},
				{"Physical", "speed of light =\ngravity =\navogadro =\nplanck =\n"},
				{"Value Lookup", "value of pi =\nvalue of speed of light =\n"},
			},
		},
		{
			Name: "Date & Time",
			Snippets: []Snippet{
				{"Current Time", "now =\ntoday =\n"},
				{"Time in City", "now in Seattle =\nnow in New York =\nnow in Kiev =\n"},
				{"Date Arithmetic", "today() =\n\\1 + 30 days =\n\\1 - 1 week =\n"},
				{"Date Difference", "19/01/22 - now =\n"},
				{"Duration Conversion", "861.5 hours in days =\n48 hours in days =\n"},
				{"Time Zone Conversion", "6:00 am Seattle in Kiev =\n11am Kiev in Seattle =\n"},
				{"Date Range", "Dec 6 till March 11 =\nJan 1 until Dec 31 =\n"},
			},
		},
		{
			Name: "Network/IP",
			Snippets: []Snippet{
				{"Subnet Info", "10.100.0.0/24 =\nsubnet info 10.100.0.0/16 =\n"},
				{"Split to Subnets", "10.100.0.0/16 / 4 subnets =\n"},
				{"Split by Hosts", "10.100.0.0/28 / 16 hosts =\n"},
				{"Subnet Mask", "mask for /24 =\nwildcard for /24 =\n"},
				{"IP in Range", "is 10.100.0.50 in 10.100.0.0/24 =\nis 192.168.1.100 in 192.168.1.0/28 =\n"},
				{"Next Subnet", "next subnet after 10.100.0.0/24 =\n"},
				{"Broadcast Address", "broadcast for 10.100.0.0/24 =\n"},
			},
		},
		{
			Name: "Unit Conversions",
			Snippets: []Snippet{
				{"Length", "5 miles in km =\n100 cm to inches =\n10 feet to meters =\n"},
				{"Weight", "10 kg in lbs =\n5 oz to grams =\n100 grams to oz =\n"},
				{"Temperature", "100 f to c =\n25 celsius to fahrenheit =\n0 kelvin to c =\n"},
				{"Volume", "5 gallons in liters =\n2 cups to ml =\n1 liter to ml =\n"},
				{"Data", "500 mb in gb =\n1 tb to gb =\n1024 kb to mb =\n"},
				{"Speed", "60 mph to kph =\n100 kph to mph =\n"},
				{"Area", "1 acre to sqft =\n100 sqm to sqft =\n1 hectare to acres =\n"},
			},
		},
		{
			Name: "Percentage",
			Snippets: []Snippet{
				{"Basic Percentage", "$100 - 20% =\n$100 + 15% =\n"},
				{"What is X% of Y", "what is 15% of 200 =\nwhat is 25% of 80 =\n"},
				{"What Percent Is", "50 is what % of 200 =\n75 is what percent of 300 =\n"},
				{"Increase/Decrease", "increase 100 by 20% =\ndecrease 500 by 15% =\n"},
				{"Percent Change", "percent change from 50 to 75 =\npercent change from 100 to 80 =\n"},
				{"Tip Calculator", "tip 20% on $85.50 =\ntip 15% on $100 =\n"},
				{"Split Bill", "$150 split 4 ways =\n$200 split 4 ways with 18% tip =\n"},
			},
		},
		{
			Name: "Financial",
			Snippets: []Snippet{
				{"Loan Payment", "loan $250000 at 6.5% for 30 years =\nloan $50000 at 4% for 5 years =\n"},
				{"Mortgage", "mortgage $350000 at 7% for 30 years =\n"},
				{"Compound Interest", "$10000 at 5% for 10 years compounded monthly =\ncompound interest $5000 at 7% for 5 years =\n"},
				{"Simple Interest", "simple interest $5000 at 3% for 2 years =\n"},
				{"Investment Growth", "invest $1000 at 7% for 20 years =\ninvest $5000 at 10% for 10 years =\n"},
			},
		},
		{
			Name: "Statistics",
			Snippets: []Snippet{
				{"Average/Mean", "avg(10, 20, 30, 40) =\nmean(1, 2, 3, 4, 5) =\n"},
				{"Median", "median(1, 2, 3, 4, 5) =\nmedian(1, 2, 3, 4, 100) =\n"},
				{"Sum & Count", "sum(10, 20, 30) =\ncount(1, 2, 3, 4, 5) =\n"},
				{"Min & Max", "min(10, 5, 20, 3) =\nmax(10, 5, 20, 3) =\n"},
				{"Standard Deviation", "stddev(2, 4, 4, 4, 5, 5, 7, 9) =\n"},
				{"Variance", "variance(2, 4, 4, 4, 5, 5, 7, 9) =\n"},
				{"Range", "range(1, 5, 10, 3) =\n"},
			},
		},
		{
			Name: "Programmer",
			Snippets: []Snippet{
				{"Bitwise AND/OR/XOR", "0xFF AND 0x0F =\n0xF0 OR 0x0F =\n0xFF XOR 0x0F =\n"},
				{"Bit Shifts", "1 << 8 =\n256 >> 4 =\n0xFF << 4 =\n"},
				{"ASCII/Char", "ascii A =\nascii a =\nchar 65 =\nchar 0x41 =\n"},
				{"ASCII Table", "ascii table =\n"},
				{"UUID Generation", "uuid =\n"},
				{"Hash Functions", "md5 hello =\nsha256 hello =\nsha1 test =\n"},
				{"Random Number", "random 1 to 100 =\nrandom 1-1000 =\n"},
			},
		},
	}
}
