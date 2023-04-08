package federation_store

var EuropeanUnionCountries = map[uint64]bool{
	54:  true, //croatia
	127: true, //luxembourg
	175: true, //portugal
	73:  true, //france
	107: true, //italy
	126: true, //lithuania
	83:  true, //greece
	154: true, //netherlands
	98:  true, //hungary
	174: true, //poland
	21:  true, //belgium
	57:  true, //czech republic
	67:  true, //estonia
	179: true, //romania
	14:  true, //austria
	72:  true, //finland
	104: true, //ireland
	196: true, //slovakia
	197: true, //slovenia
	56:  true, //cyprus
	80:  true, //germany
	120: true, //latvia
	202: true, //spain
	208: true, //sweden
	33:  true, //bulgaria
	58:  true, //denmark
	135: true, //malta
}
var EuropeanCountries = map[uint64]bool{
	180: true, //russia
	209: true, //switzerland
	228: true, //united kingdom
	2:   true, //albania
	5:   true, //andorra
	57:  true, //czech republic
	163: true, //norway
	208: true, //sweden
	95:  true, //holy see
	135: true, //malta
	144: true, //monaco
	196: true, //slovakia
	179: true, //romania
	80:  true, //germany
	146: true, //montenegro
	192: true, //serbia
	197: true, //slovenia
	33:  true, //bulgaria
	54:  true, //croatia
	58:  true, //denmark
	67:  true, //estonia
	14:  true, //austria
	20:  true, //belarus
	143: true, //moldova, republic of
	21:  true, //belgium
	72:  true, //finland
	73:  true, //france
	202: true, //spain
	104: true, //ireland
	107: true, //italy
	126: true, //lithuania
	127: true, //luxembourg
	1:   true, //åland islands
	27:  true, //bosnia and herzegovina
	83:  true, //greece
	98:  true, //hungary
	154: true, //netherlands
	129: true, //north macedonia
	188: true, //san marino
	120: true, //latvia
	174: true, //poland
}
var MiddleEastCountries = map[uint64]bool{
	221: true, //turkey
	240: true, //yemen
	56:  true, //cyprus
	103: true, //iraq
	111: true, //jordan
	121: true, //lebanon
	17:  true, //bahrain
	227: true, //united arab emirates
	102: true, //iran
	167: true, //palestine
	177: true, //qatar
	190: true, //saudi arabia
	210: true, //syria
	63:  true, //egypt
	106: true, //israel
	117: true, //kuwait
	164: true, //oman
}
var AfricanCountries = map[uint64]bool{
	159: true, //nigeria
	194: true, //sierra leone
	123: true, //liberia
	220: true, //tunisia
	151: true, //namibia
	189: true, //sao tome and principe
	182: true, //saint helena
	130: true, //madagascar
	193: true, //seychelles
	39:  true, //cape verde
	6:   true, //angola
	207: true, //eswatini
	41:  true, //central african republic
	49:  true, //congo
	178: true, //reunion
	204: true, //sudan
	42:  true, //chad
	78:  true, //gambia
	66:  true, //eritrea
	37:  true, //cameroon
	138: true, //mauritania
	90:  true, //guinea
	53:  true, //cote d"ivoire
	28:  true, //botswana
	200: true, //south africa
	158: true, //niger
	181: true, //rwanda
	199: true, //somalia
	241: true, //zambia
	77:  true, //gabon
	34:  true, //burkina faso
	124: true, //libya
	148: true, //morocco
	131: true, //malawi
	191: true, //senegal
	139: true, //mauritius
	134: true, //mali
	48:  true, //comoros
	59:  true, //djibouti
	113: true, //kenya
	213: true, //tanzania
	23:  true, //benin
	35:  true, //burundi
	68:  true, //ethiopia
	149: true, //mozambique
	242: true, //zimbabwe
	225: true, //uganda
	50:  true, //congo, democratic republic
	65:  true, //equatorial guinea
	122: true, //lesotho
	81:  true, //ghana
	91:  true, //guinea-bissau
	3:   true, //algeria
	239: true, //western sahara
	216: true, //togo
}
var OceaniaCountries = map[uint64]bool{
	238: true, //wallis and futuna
	71:  true, //fiji
	87:  true, //guam
	156: true, //new zealand
	224: true, //tuvalu
	233: true, //vanuatu
	160: true, //niue
	166: true, //palau
	187: true, //samoa
	4:   true, //american samoa
	51:  true, //cook islands
	75:  true, //french polynesia
	114: true, //kiribati
	136: true, //marshall islands
	198: true, //solomon islands
	169: true, //papua new guinea
	173: true, //pitcairn
	162: true, //northern mariana islands
	217: true, //tokelau
	218: true, //tonga
	13:  true, //australia
	142: true, //micronesia
	152: true, //nauru
	155: true, //new caledonia
	161: true, //norfolk island
}
var NorthAmericaCountries = map[uint64]bool{
	253: true, //curaçao
	86:  true, //guadeloupe
	168: true, //panama
	176: true, //puerto rico
	183: true, //saint kitts and nevis
	184: true, //saint lucia
	7:   true, //anguilla
	12:  true, //aruba
	108: true, //jamaica
	237: true, //virgin islands, u.s.
	9:   true, //antigua and barbuda
	24:  true, //bermuda
	40:  true, //cayman islands
	96:  true, //honduras
	16:  true, //bahamas
	22:  true, //belize
	38:  true, //canada
	88:  true, //guatemala
	185: true, //saint pierre and miquelon
	223: true, //turks and caicos islands
	85:  true, //grenada
	229: true, //united states
	236: true, //virgin islands, british
	64:  true, //el salvador
	93:  true, //haiti
	147: true, //montserrat
	157: true, //nicaragua
	186: true, //saint vincent and the grenadines
	219: true, //trinidad and tobago
	19:  true, //barbados
	52:  true, //costa rica
	60:  true, //dominica
	137: true, //martinique
	61:  true, //dominican republic
	84:  true, //greenland
	141: true, //mexico
}
var SouthAmericaCountries = map[uint64]bool{
	10:  true, //argentina
	171: true, //peru
	231: true, //uruguay
	201: true, //south georgia
	30:  true, //brazil
	43:  true, //chile
	69:  true, //falkland islands (malvinas)
	234: true, //venezuela
	26:  true, //bolivia
	92:  true, //guyana
	205: true, //suriname
	170: true, //paraguay
	47:  true, //colombia
	62:  true, //ecuador
	74:  true, //french guiana
}

var AllCountries = map[uint64]map[uint64]bool{
	245: NorthAmericaCountries,
	246: SouthAmericaCountries,
	248: EuropeanCountries,
	247: EuropeanUnionCountries,
	249: MiddleEastCountries,
	250: OceaniaCountries,
	251: AfricanCountries,
}
