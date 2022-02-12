package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/tidwall/gjson"
)

var InstanceName string = os.Getenv("InstanceName")

func GetPublicIP(svc *lightsail.Lightsail) string {
	getInstance, _ := svc.GetInstance(&lightsail.GetInstanceInput{
		InstanceName: &InstanceName,
	})
	return *getInstance.Instance.PublicIpAddress
}

func GetStatus(svc *lightsail.Lightsail) string {
	getInstanceState, _ := svc.GetInstanceState(&lightsail.GetInstanceStateInput{
		InstanceName: &InstanceName,
	})
	return ParseStatus(getInstanceState.State.String())
}

func GetTotalNetworkPerMonth(svc *lightsail.Lightsail) string {

	var MetricOut string = lightsail.MetricNameNetworkOut
	var MetricIn string = lightsail.MetricNameNetworkIn
	var Period int64 = 2592000
	var StartTime time.Time = BeginningOfMonth(time.Now())
	var EndTime time.Time = time.Now()
	var Unit string = lightsail.MetricUnitBytes

	//Dumb ways to die
	Statistics := []*string{}
	input := "Sum"
	Statistics = append(Statistics, &input)

	trafficIn, _ := svc.GetInstanceMetricData(&lightsail.GetInstanceMetricDataInput{
		InstanceName: &InstanceName,
		MetricName:   &MetricIn,
		Period:       &Period,
		StartTime:    &StartTime,
		EndTime:      &EndTime,
		Unit:         &Unit,
		Statistics:   Statistics,
	})
	trafficOut, _ := svc.GetInstanceMetricData(&lightsail.GetInstanceMetricDataInput{
		InstanceName: &InstanceName,
		MetricName:   &MetricOut,
		Period:       &Period,
		StartTime:    &StartTime,
		EndTime:      &EndTime,
		Unit:         &Unit,
		Statistics:   Statistics,
	})
	return ParseTrafficValue(trafficIn.MetricData[0].String()) + "/" + ParseTrafficValue(trafficOut.MetricData[0].String())
}

func BeginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

func regJsonData(data string) string {
	reg := regexp.MustCompile("([a-zA-Z]\\w*):")
	regStr := reg.ReplaceAllString(data, `"$1":`)
	return regStr
}

func ParseTrafficValue(data string) string {
	val, _ := strconv.ParseFloat(gjson.Get(regJsonData(data), "Sum").Raw, 32)
	switch {
	case val > 1024 && val < 1048576:
		return Traffic2String(val/1024) + "KB" //kilobyte
	case val > 1048576 && val < 1073741824:
		return Traffic2String(val/1048576) + "MB" //megabyte
	case val > 1073741824:
		return Traffic2String(val/1073741824) + "GB" //gigabyte
	default:
		return Traffic2String(val) + "B" //byte
	}
}

func Traffic2String(val float64) string {
	return fmt.Sprintf("%.2f", math.Round(val*(math.Pow10(2)))/math.Pow10(2))
}

func ParseStatus(data string) string {
	return gjson.Get(regJsonData(data), "Name").String()
}
