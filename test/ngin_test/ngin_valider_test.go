package ngintest

import (
	"testing"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/ngin"
	"github.com/niexqc/nlibs/ntools"
)

func TestValiderBase(t *testing.T) {
	valider := ngin.NewNValider("json", "zhdesc")
	// 接收账套数据
	type RecvTaskZhangtaoVo struct {
		Id         int64             `db:"id" json:"id" zhdesc:""`
		DataKey    string            `db:"data_key" json:"dataKey" zhdesc:"账套唯一标识" binding:"required,max=36" `
		Bm         string            `db:"bm" json:"bm" zhdesc:"账套编码" binding:"required,max=20" `
		Mc         string            `db:"mc" json:"mc" zhdesc:"账套名称" binding:"required,max=255" `
		Tyshxydm   string            `db:"tyshxydm" json:"tyshxydm" zhdesc:"账套所属组织的统一社会信用代码" binding:"required,len=18" `
		FiscalYear int               `db:"fiscal_year" json:"fiscalYear" zhdesc:"所属年份" binding:"required,gt=2023" `
		QyRq       string            `db:"qy_rq" json:"qyRq" zhdesc:"账套启用期间" binding:"required,len=6"`
		TyRq       sqlext.NullString `db:"ty_rq" json:"tyRq" zhdesc:"账套停用期间" binding:"omitempty,len=6" `
		GdzcQyRq   sqlext.NullString `db:"gdzc_qy_rq" json:"gdzcQyRq" zhdesc:"固定资产启用日期" binding:"omitempty,len=6"`
		GdzcTyRq   sqlext.NullString `db:"gdzc_ty_rq" json:"gdzcTyRq" zhdesc:"固定资产停用日期" binding:"omitempty,len=6"`
		PzhWs      int               `db:"pzh_ws" json:"pzhWs" zhdesc:"凭证号显示位数" binding:"required,oneof=1 2 3 4 5 6 7 8 9"`
		KjQj       string            `db:"kj_qj" json:"kjQj" zhdesc:"当前会计期间" binding:"required,len=6"`
		BizId      string            `db:"biz_id" json:"bizId" zhdesc:"业务唯一id" binding:"required,max=32" `
		AreaCode   string            `db:"area_code" json:"areaCode" zhdesc:"行政区划编码" binding:"required,max=15" `
	}

	ztVo := &RecvTaskZhangtaoVo{DataKey: "111", Bm: "111111111111111111111111111", TyRq: sqlext.NewNullString(true, "122221"), Tyshxydm: "111", FiscalYear: 2023, PzhWs: 12}
	err := valider.ValidStrct(ztVo)
	ntools.TestErrNotNil(t, "TestValiderBase", err)
	err = valider.TransErr2ZhErr(err)

	ntools.TestStrContains(t, "TestValiderBase", "账套编码[bm]", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "账套名称[mc]为必填字段", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "统一社会信用代码[tyshxydm]长度必须是18个字符", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "所属年份[fiscalYear]", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "账套启用期间[qyRq]为必填字段", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "凭证号显示位数[pzhWs]", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "当前会计期间[kjQj]为必填字段", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "业务唯一id[bizId]为必填字段", err.Error())
	ntools.TestStrContains(t, "TestValiderBase", "行政区划编码[areaCode]为必填字段", err.Error())
}

func TestValiderOneofZhc(t *testing.T) {
	valider := ngin.NewNValider("json", "zhdesc")
	// 接收账套数据
	type RecvTaskZhangtaoVo struct {
		PzhWs string `db:"pzh_ws" json:"pzhWs" zhdesc:"凭证号显示位数" binding:"required,oneof=一 二 三"`
	}

	ztVo := &RecvTaskZhangtaoVo{PzhWs: "四"}
	err := valider.ValidStrct(ztVo)
	ntools.TestErrNotNil(t, "TestValiderBase", err)
	err = valider.TransErr2ZhErr(err)

	ntools.TestStrContains(t, "TestValiderBase", "凭证号显示位数[bm]", err.Error())
}
