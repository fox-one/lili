package house

import (
	"errors"
	"fmt"
	"time"

	. "github.com/bearyinnovative/lili/model"
	. "github.com/bearyinnovative/lili/util"
)

const (
	pageCount = 100
)

var cityIdMap map[string]int

func init() {
	cityIdMap = map[string]int{
		"北京":  110000,
		"上海":  310000,
		"广州":  440100,
		"深圳":  440300,
		"天津":  120000,
		"成都":  510100,
		"南京":  320100,
		"杭州":  330100,
		"青岛":  370200,
		"大连":  210200,
		"厦门":  350200,
		"武汉":  420100,
		"重庆":  500000,
		"长沙":  430100,
		"济南":  370101,
		"佛山":  440600,
		"东莞":  441900,
		"西安":  610100,
		"石家庄": 130100,
		"烟台":  370600,
		"中山":  442000,
		"珠海":  440400,
		"沈阳":  210100,
		"合肥":  340100,
	}
}

type BaseHouseDeal struct {
	cityName      string
	cityShortName string
}

func (c *BaseHouseDeal) Name() string {
	return "house-deal-" + c.cityName
}

func (c *BaseHouseDeal) Interval() time.Duration {
	return time.Hour * 24
}

func (c *BaseHouseDeal) Notifiers() []NotifierType {
	return houseNotifiers
}

func (c *BaseHouseDeal) Fetch() (results []*Item, err error) {
	cityId, err := getCityIdFromName(c.cityName)
	if LogIfErr(err) {
		return
	}

	stop := false
	offset := 0
	limit := pageCount

	for !stop {
		dealItems, err := FetchDeals(cityId, offset, limit)
		if LogIfErr(err) {
			break
		}

		for _, di := range dealItems {
			created, err := UpsertDeal(di)

			if LogIfErr(err) {
				stop = true
				break
			}

			if !created {
				stop = true
				continue
			}

			// {"title" : "南岭花园 1室1厅 29.24㎡", "price" : 648000, "pricehide" : "6*", "deschide" : "近30天内成交", "unitprice" : 22162, "signdate" : "2017.11.04", "signtimestamp" : 1509788751, "signsource" : "链家成交", "orientation" : "南", "floorstate" : "低楼层/1层", "buildingfinishyear" : 1994, "decoration" : "简装", "buildingtype" : "板楼", "requirelogin" : 0, "fetchedat" : ISODate("2017-11-19T04:29:44.621Z") }
			createdAt := time.Unix(int64(di.SignTimestamp), 0)
			ref := fmt.Sprintf("https://%s.lianjia.com/chengjiao/%s.html", c.cityShortName, di.HouseCode)
			item := &Item{
				Name:       c.Name(),
				Identifier: c.Name() + "-" + di.HouseCode,
				// 南岭花园 1室1厅 29.24㎡ 南 | 简装 | 低楼层/1层 | 板楼 总价: 648000 单价: 22162 成交时间 2017.11.04
				Desc: fmt.Sprintf("**NEW DEAL** %s %s %s | %s | %s | %s 总价: %.1f万 单价: %.4f万 成交时间: %s [Link](%s)",
					c.Name(), di.Title, di.Orientation, di.Decoration, di.FloorState, di.BuildingType, float64(di.Price)/10000.0, float64(di.UnitPrice)/10000.0, di.SignDate, ref),
				Ref:     ref,
				Created: createdAt,
				Images:  []string{di.CoverPic},
			}
			results = append(results, item)
		}

		offset += limit
	}

	return
}

func getCityIdFromName(name string) (int, error) {
	id, present := cityIdMap[name]
	if !present {
		return -1, errors.New("can't find city id for: " + name)
	}

	return id, nil
}
