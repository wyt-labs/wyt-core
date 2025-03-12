package coincarp

type Category struct {
	Code string `bson:"code"`
	Name string `bson:"name"`
}

type Investor struct {
	InvestorLogo string `bson:"investorlogo"`
	InvestorCode string `bson:"investorcode"`
	InvestorName string `bson:"investorname"`
}

type ReadDB struct {
	ID            string     `bson:"_id"`
	ProjectCode   string     `bson:"projectcode"`
	ProjectName   string     `bson:"projectname"`
	Logo          string     `bson:"logo"`
	CategoryList  []Category `bson:"categorylist"`
	FundCode      string     `bson:"fundcode"`
	FundStageCode string     `bson:"fundstagecode"`
	FundStageName string     `bson:"fundstagename"`
	FundAmount    uint       `bson:"fundamount"`
	Valulation    uint       `bson:"valulation"`
	FundDate      int64      `bson:"funddate"`
	InvestorCodes string     `bson:"investorcodes"`
	InvestorNames string     `bson:"investornames"`
	InvestorLogos string     `bson:"investorlogos"`
	InvestorCount int        `bson:"investorcount"`
	InvestorList  []Investor `bson:"investorlist"`
}
