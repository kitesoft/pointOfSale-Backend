package sql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "stock/common/logger"
	. "stock/entities"
	"strconv"
	"stock/entities/responses"
	"time"
)

const stTableSaleBasket = `CREATE TABLE IF NOT EXISTS %s.sale_basket (
						  id             INT AUTO_INCREMENT PRIMARY KEY,
						  creation_date  INT     NOT NULL DEFAULT 0,
						  items			 TEXT,
						  user_id 		 INT 	DEFAULT 1,
						  total_price	 FLOAT	DEFAULT 0,
						  total_discount FLOAT 	DEFAULT 0,
						  FOREIGN KEY (user_id) REFERENCES %s.user (id) ON DELETE CASCADE ON UPDATE CASCADE	
						)ENGINE=InnoDB DEFAULT CHARSET=utf8;`



const stSelectSaleBasketById = `SELECT id,creation_date,items,user_id,total_price,total_discount FROM %s.sale_basket
									 WHERE id=?`

const stInsertSaleBasket = `INSERT INTO %s.sale_basket (creation_date,items,user_id,total_price,total_discount)
							VALUES (?,?,?,?,?)`

const stUpdateSaleBasketById = `UPDATE %s.sale_basket SET creation_date=?,items=?,user_id=?,total_price=?,total_discount=?
								WHERE id=?`

const stDeleteSaleBasketById = `DELETE FROM %s.sale_basket WHERE id=?`

type SaleBasketRepo struct {}

var sl *SaleBasketRepo
var qSelectSaleBasketById,qInsertSaleBasket,qUpdateSaleBasketById,qDeleteSaleBasketById *sql.Stmt

func GetSaleBasketRepo() *SaleBasketRepo{
	if sl == nil {
		sl = &SaleBasketRepo{}

		var err error
		if _, err = DB.Exec(ss(stTableSaleBasket)); err != nil {
			LogError(err)
		}

		qSelectSaleBasketById, err = DB.Prepare(s(stSelectSaleBasketById))
		if err != nil {
			LogError(err)
		}

		qInsertSaleBasket, err = DB.Prepare(s(stInsertSaleBasket))
		if err != nil {
			LogError(err)
		}

		qUpdateSaleBasketById, err = DB.Prepare(s(stUpdateSaleBasketById))
		if err != nil {
			LogError(err)
		}
		qDeleteSaleBasketById, err = DB.Prepare(s(stDeleteSaleBasketById))
		if err != nil {
			LogError(err)
		}
	}

	return sl
}

func (sl *SaleBasketRepo) SelectSaleBasketById(id int)(*SaleBasket,error){
	p := &SaleBasket{}
	row := qSelectSaleBasketById.QueryRow(id)
	err := row.Scan(&p.Id,&p.CreationDate,&p.ItemsStr,&p.UserId,&p.TotalPrice,&p.TotalDiscount)
	if err != nil{
		LogError(err)
		return nil, err
	}
	return p,nil
}


func (sl *SaleBasketRepo) InsertSaleBasket(p *SaleBasket)(error){

	timeNOW := int(time.Now().Unix())
	result,err := qInsertSaleBasket.Exec(timeNOW,p.ItemsStr,p.UserId,p.TotalPrice,p.TotalDiscount)
	if err != nil{
		LogError(err)
		return err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		LogError(err)
		return err
	}
	p.Id = int(lastId)

	return nil
}

func (sl *SaleBasketRepo) UpdateSaleBasketById(p *SaleBasket, IdToUpdate int)(error){

	timeNOW := int(time.Now().Unix())
	_,err := qUpdateSaleBasketById.Exec(timeNOW,p.ItemsStr,p.UserId,p.TotalPrice,p.TotalDiscount,IdToUpdate)
	if err != nil{
		LogError(err)
		return err
	}

	return nil
}

func (sl *SaleBasketRepo) DeleteSaleBasketById(Id int)(error){

	_,err := qDeleteSaleBasketById.Exec(Id)
	if err != nil{
		LogError(err)
		return err
	}

	return nil
}

func (sl *SaleBasketRepo) DeleteSaleBaskets(ids []int)(error){


	stDelete := `DELETE FROM %s.sale_basket WHERE id in (`

	for k,v := range ids{
		stDelete += strconv.FormatInt(int64(v),10)
		if k < len(ids)-1{
			stDelete+=`,`
		}
	}
	stDelete += `)`

	stDelete = s(stDelete)
	LogDebug(stDelete)

	qDelete,err := DB.Prepare(stDelete)
	defer qDelete.Close()
	if err != nil{
		LogDebug(err)
		return err
	}

	_,err = qDelete.Exec()
	if err != nil{
		LogError(err)
		return err
	}

	return nil
}

func (sl *SaleBasketRepo) SelectSaleBaskets(timeInterval []int,userId int,orderBy,orderAs string,pageNumber, pageSize int) (*responses.SaleBasketResponse,  error) {

	response := &responses.SaleBasketResponse{}
	items := []*SaleBasket{}

	var timeAvail bool
	var userAvail bool

	var orderByAvail bool
	var pageNumberAvail bool
	var pageSizeAvail bool


	if len(timeInterval) > 0{
		timeAvail = true
	}

	if userId > 0 {
		userAvail = true
	}

	if len(orderBy) != 0{
		if orderBy != "id" {
			orderByAvail = true
		}
	}
	if pageNumber > 0 {
		pageNumberAvail = true
	}
	if pageSize > 0 {
		pageSizeAvail = true
	}

	stSelect := `SELECT s.id,s.creation_date,s.items,s.user_id,u.name, s.total_price, s.total_discount 
						FROM %s.sale_basket AS s
						JOIN %s.user AS u ON u.id=s.user_id`
	stCount := `SELECT COUNT(*) FROM %s.sale_basket AS s
						JOIN %s.user AS u ON u.id=s.user_id`

	stSelect = ss(stSelect)
	stCount = ss(stCount)

	filter := ``

	if  timeAvail || userAvail{
		filter += " WHERE "


		if timeAvail{
			filter += " s.creation_date > " + strconv.FormatInt(int64(timeInterval[0]),10)
			filter += " AND s.creation_date < " + strconv.FormatInt(int64(timeInterval[1]),10)

			if userAvail{
				filter += ` AND `
			}
		}

		if userAvail{
			filter += ` s.user_id = ` + strconv.FormatInt(int64(userId),10)
		}


	}

	stSelect += filter
	stCount += filter

	stSelect += ` ORDER BY `
	if orderByAvail {
		stSelect +=  orderBy
	}else{
		stSelect += ` s.id `
	}
	if orderAs == "asc"{
		stSelect += ` ASC `
	}else{
		stSelect += ` DESC `
	}
	if pageNumberAvail && pageSizeAvail {
		offset := strconv.FormatInt(int64((pageNumber-1)*pageSize),10)
		pageSizeStr := strconv.FormatInt(int64(pageSize),10)
		stSelect += ` LIMIT ` + offset + `,` + pageSizeStr
	}

	LogDebug(stSelect)

	qSelect, err := DB.Prepare(stSelect)
	defer qSelect.Close()

	if err != nil{
		LogError(err)
		return nil, err
	}

	rows, err := qSelect.Query()
	if err != nil{
		LogError(err)
		return nil, err
	}

	for rows.Next(){
		p := &SaleBasket{}
		err = rows.Scan(&p.Id,&p.CreationDate,&p.ItemsStr,&p.UserId,&p.UserName,&p.TotalPrice,&p.TotalDiscount)
		if err != nil {
			LogError(err)
		}
		items = append(items, p)
	}

	response.Items = items

	qCount, err := DB.Prepare(stCount)
	defer qCount.Close()
	if err != nil{
		LogError(err)
		return nil, err
	}
	count := qCount.QueryRow()
	if err != nil{
		LogError(err)
		return nil, err
	}
	count.Scan(&response.Count)

	return response,nil
}

func (sl *SaleBasketRepo) Close() {
	qSelectSaleBasketById.Close()
	qInsertSaleBasket.Close()
	qUpdateSaleBasketById.Close()
	qDeleteSaleBasketById.Close()
}