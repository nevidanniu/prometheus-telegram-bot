package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"math"
	"strconv"
	"strings"
	"time"
)

var funcMap = template.FuncMap{
	"delimiter":              delimiter,
	"str_FormatDate":         str_FormatDate,
	"str_UpperCase":          strings.ToUpper,
	"str_LowerCase":          strings.ToLower,
	"str_Title":              strings.Title,
	"str_FormatFloat":        str_FormatFloat,
	"str_Format_Byte":        str_Format_Byte,
	"str_Format_MeasureUnit": str_Format_MeasureUnit,
	"HasKey":                 HasKey,
}

/**
 * Subdivideb by 1024
 */
const (
	Kb = iota
	Mb
	Gb
	Tb
	Pb
	Eb
	Zb
	Yb
	Information_Size_MAX
)

/**
 * Subdivided by 10000
 */
const (
	K = iota
	M
	G
	T
	P
	E
	Z
	Y
	Scale_Size_MAX
)

func RoundPrec(x float64, prec int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow * sign
}

/******************************************************************************
 *
 *          Function for formatting template
 *
 ******************************************************************************/

func delimiter(_ string) string {
	return "[delimiter]"
}

func str_Format_MeasureUnit(MeasureUnit string, value string) string {
	var RetStr string
	cfg.SplitChart = "|"
	MeasureUnit = strings.TrimSpace(MeasureUnit) // Remove space
	SplittedMUnit := strings.SplitN(MeasureUnit, cfg.SplitChart, 3)

	Initial := 0
	// If is declared third part of array, then Measure unit start from just scaled measure unit.
	// Example Kg is Kilo g, but all people use Kg not g, then you will put here 3 Kilo. Bot strart convert from here.
	if len(SplittedMUnit) > 2 {
		tmp, err := strconv.ParseInt(SplittedMUnit[2], 10, 8)
		if err != nil {
			log.Println("Could not convert value to int")
			if !cfg.Debug {
				// If is running in production leave daemon live. else here will die with log error.
				return "" // Break execution and return void string, bot will log somethink
			}
		}
		Initial = int(tmp)
	}

	switch SplittedMUnit[0] {
	case "kb":
		RetStr = str_Format_Byte(value, Initial)
	case "s":
		RetStr = str_Format_Scale(value, Initial)
	case "f":
		RetStr = str_FormatFloat(value)
	case "i":
		RetStr = str_FormatInt(value)
	default:
		RetStr = str_FormatInt(value)
	}

	if len(SplittedMUnit) > 1 {
		RetStr += SplittedMUnit[1]
	}

	return RetStr
}

// Scale number for It measure unit
func str_Format_Byte(in string, j1 int) string {
	var str_Size string

	f, err := strconv.ParseFloat(in, 64)

	if err != nil {
		panic(err)
	}

	for j1 = 0; j1 < (Information_Size_MAX + 1); j1++ {

		if j1 >= Information_Size_MAX {
			str_Size = "Yb"
			break
		} else if f > 1024 {
			f /= 1024.0
		} else {

			switch j1 {
			case Kb:
				str_Size = "Kb"
			case Mb:
				str_Size = "Mb"
			case Gb:
				str_Size = "Gb"
			case Tb:
				str_Size = "Tb"
			case Pb:
				str_Size = "Pb"
			case Eb:
				str_Size = "Eb"
			case Zb:
				str_Size = "Zb"
			case Yb:
				str_Size = "Yb"
			}
			break
		}
	}

	str_fl := strconv.FormatFloat(f, 'f', 2, 64)
	return fmt.Sprintf("%s %s", str_fl, str_Size)
}

// Format number for fisics measure unit
func str_Format_Scale(in string, j1 int) string {
	var str_Size string

	f, err := strconv.ParseFloat(in, 64)

	if err != nil {
		panic(err)
	}

	for j1 = 0; j1 < (Scale_Size_MAX + 1); j1++ {

		if j1 >= Scale_Size_MAX {
			str_Size = "Y"
			break
		} else if f > 1000 {
			f /= 1000.0
		} else {
			switch j1 {
			case K:
				str_Size = "K"
			case M:
				str_Size = "M"
			case G:
				str_Size = "G"
			case T:
				str_Size = "T"
			case P:
				str_Size = "P"
			case E:
				str_Size = "E"
			case Z:
				str_Size = "Z"
			case Y:
				str_Size = "Y"
			default:
				str_Size = "Y"
			}
			break
		}
	}

	str_fl := strconv.FormatFloat(f, 'f', 2, 64)
	return fmt.Sprintf("%s %s", str_fl, str_Size)
}

func str_FormatInt(i string) string {
	v, _ := strconv.ParseInt(i, 10, 64)
	val := strconv.FormatInt(v, 10)
	return val
}

func str_FormatFloat(f string) string {
	v, _ := strconv.ParseFloat(f, 64)
	v = RoundPrec(v, 2)
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func str_FormatDate(toformat string) string {

	// Error handling
	if cfg.TimeZone == "" {
		log.Println("template_time_zone is not set, if you use template and `str_FormatDate` func is required")
		panic(nil)
	}

	if cfg.TimeOutFormat == "" {
		log.Println("template_time_outdata param is not set, if you use template and `str_FormatDate` func is required")
		panic(nil)
	}

	t, err := time.Parse(time.RFC3339Nano, toformat)

	if err != nil {
		fmt.Println(err)
	}

	loc, _ := time.LoadLocation(cfg.TimeZone)

	return t.In(loc).Format(cfg.TimeOutFormat)
}

func HasKey(dict map[string]interface{}, key_search string) bool {
	if _, ok := dict[key_search]; ok {
		return true
	}
	return false
}
