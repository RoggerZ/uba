package request

import "encoding/json"

// GmRoleModel
type GmRoleModel struct {
	ID          int      `json:"id" db:"id"`
	RoleName    string   `json:"name" db:"role_name"`
	Description string   `json:"description" db:"description"`
	RoleList    string   `json:"routes" db:"role_list"`
	Api         []string `json:"api"`
}

type AnalysisFilter struct {
	FilterType string `json:"filterType"`
	Filts      []struct {
		FilterType string `json:"filterType"`
		Filts      []struct {
			ColumnName string      `json:"columnName"`
			Comparator string      `json:"comparator"`
			FilterType string      `json:"filterType"`
			Ftv        interface{} `json:"ftv"`
		} `json:"filts,omitempty"`
		Relation   string      `json:"relation,omitempty"`
		ColumnName string      `json:"columnName,omitempty"`
		Comparator string      `json:"comparator,omitempty"`
		Ftv        interface{} `json:"ftv,omitempty"`
	} `json:"filts"`
	Relation string `json:"relation"`
}

type Zhibiao struct {
	EventName        string         `json:"eventName"`
	EventNameDisplay string         `json:"eventNameDisplay"`
	Relation         AnalysisFilter `json:"relation"`
}

type FunnelReqData struct {
	UserGroup         []int          `json:"userGroup"`
	ZhibiaoArr        []Zhibiao      `json:"zhibiaoArr"`
	WhereFilter       AnalysisFilter `json:"whereFilter"`
	WindowTime        int            `json:"windowTime"`
	WindowTimeFormat  string         `json:"windowTimeFormat"`
	ResultTimeFormat  string         `json:"resultTimeFormat"`
	Date              []string       `json:"date"`
	Appid             int            `json:"appid"`
	WhereFilterByUser AnalysisFilter `json:"whereFilterByUser"`
	GroupBy           []string       `json:"groupBy"`
}

type TraceReqData struct {
	EventNames        []string       `json:"eventNames"`
	UserGroup         []int          `json:"userGroup"`
	ZhibiaoArr        []Zhibiao      `json:"zhibiaoArr"`
	WhereFilter       AnalysisFilter `json:"whereFilter"`
	WindowTime        int            `json:"windowTime"`
	WindowTimeFormat  string         `json:"windowTimeFormat"`
	Date              []string       `json:"date"`
	Appid             int            `json:"appid"`
	WhereFilterByUser AnalysisFilter `json:"whereFilterByUser"`
	GroupBy           []string       `json:"groupBy"`
}

type RetentionReqData struct {
	UserGroup  []int `json:"userGroup"`
	ZhibiaoArr []struct {
		EventName        string         `json:"eventName"`
		EventNameDisplay string         `json:"eventNameDisplay"`
		Relation         AnalysisFilter `json:"relation"`
	} `json:"zhibiaoArr"`
	WhereFilter       AnalysisFilter `json:"whereFilter"`
	WindowTime        int            `json:"windowTime"`
	WindowTimeFormat  string         `json:"windowTimeFormat"`
	Date              []string       `json:"date"`
	Appid             int            `json:"appid"`
	WhereFilterByUser AnalysisFilter `json:"whereFilterByUser"`
	GroupBy           []string       `json:"groupBy"`
}

type LTVReqData struct {
	UserGroup  []int `json:"userGroup"`
	ZhibiaoArr []struct {
		EventName        string         `json:"eventName"`
		EventNameDisplay string         `json:"eventNameDisplay"`
		Relation         AnalysisFilter `json:"relation"`
		ValueField       string         `json:"valueField"` // Field to sum (e.g. "pay_amount")
	} `json:"zhibiaoArr"`
	WhereFilter       AnalysisFilter `json:"whereFilter"`
	WindowTime        int            `json:"windowTime"`
	WindowTimeFormat  string         `json:"windowTimeFormat"`
	Date              []string       `json:"date"`
	Appid             int            `json:"appid"`
	WhereFilterByUser AnalysisFilter `json:"whereFilterByUser"`
	GroupBy           []string       `json:"groupBy"`
}

type AttributionEvent struct {
	EventName        string         `json:"eventName"`
	EventNameDisplay string         `json:"eventNameDisplay"`
	Relation         AnalysisFilter `json:"relation"`
	ValueField       string         `json:"valueField,omitempty"`
	LinkField        string         `json:"linkField,omitempty"`
}

type AttributionReqData struct {
	UserGroup               []int            `json:"userGroup"`
	ConversionEvent         AttributionEvent `json:"conversionEvent"`
	ForwardEvent            AttributionEvent `json:"forwardEvent"`
	TouchArr                []Zhibiao        `json:"touchArr"`
	WhereFilter             AnalysisFilter   `json:"whereFilter"`
	WhereFilterByUser       AnalysisFilter   `json:"whereFilterByUser"`
	GroupBy                 []string         `json:"groupBy"`
	WindowTime              int              `json:"windowTime"`
	WindowTimeFormat        string           `json:"windowTimeFormat"`
	ConversionTimeFormat    string           `json:"conversionTimeFormat"`
	IncludeDirectConversion bool             `json:"includeDirectConversion"`
	Date                    []string         `json:"date"`
	Appid                   int              `json:"appid"`
	AttributionModel        string           `json:"attributionModel"`
}

type FormulaDimension struct {
	SelectAttr []string       `json:"selectAttr"`
	EventName  string         `json:"eventName"`
	Relation   AnalysisFilter `json:"relation"`
}

type ChannelCostAddReq struct {
	AppID    int     `json:"appid"`
	Channel  string  `json:"channel"`
	CostDate string  `json:"costDate"`
	Cost     float64 `json:"cost"`
}

type ChannelCostUpdateReq struct {
	ID   int64   `json:"id"`
	Cost float64 `json:"cost"`
}

type ChannelCostDeleteReq struct {
	ID int64 `json:"id"`
}

type ChannelCostListReq struct {
	AppID     int    `json:"appid"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type EventZhibiao struct {
	SelectAttr        []string         `json:"selectAttr,omitempty"`
	Typ               int              `json:"typ"`
	EventName         string           `json:"eventName,omitempty"`
	EventNameDisplay  string           `json:"eventNameDisplay"`
	Relation          AnalysisFilter   `json:"relation,omitempty"`
	ScaleType         string           `json:"scaleType,omitempty"`
	Operate           string           `json:"operate,omitempty"`
	One               FormulaDimension `json:"one,omitempty"`
	Two               FormulaDimension `json:"two,omitempty"`
	DivisorNoGrouping bool             `json:"divisor_no_grouping"`
}

type EventReqData struct {
	UserGroup         []int          `json:"userGroup"`
	ZhibiaoArr        []EventZhibiao `json:"zhibiaoArr"`
	GroupBy           []string       `json:"groupBy"`
	WhereFilter       AnalysisFilter `json:"whereFilter"`
	WhereFilterByUser AnalysisFilter `json:"whereFilterByUser"`
	Date              []string       `json:"date"`
	WindowTimeFormat  string         `json:"windowTimeFormat"`
	Appid             int            `json:"appid"`
}

type LeaderboardMetric struct {
	MetricType   string `json:"metricType"`
	ValueField   string `json:"valueField"`
	DisplayName  string `json:"displayName"`
	SuccessValue string `json:"successValue"`
}

type LeaderboardReqData struct {
	UserGroup         []int             `json:"userGroup"`
	WhereFilter       AnalysisFilter    `json:"whereFilter"`
	WhereFilterByUser AnalysisFilter    `json:"whereFilterByUser"`
	Appid             int               `json:"appid"`
	EventName         string            `json:"eventName"`
	EventNameDisplay  string            `json:"eventNameDisplay"`
	Metric            LeaderboardMetric `json:"metric"`
	GroupBy           []string          `json:"groupBy"`
	Date              []string          `json:"date"`
	CompareDate       []string          `json:"compareDate"`
	RankingMode       string            `json:"rankingMode"`
	SortBy            string            `json:"sortBy"`
	SortOrder         string            `json:"sortOrder"`
	TopN              int               `json:"topN"`
	IncludeOthers     bool              `json:"includeOthers"`
	ExcludeEmpty      bool              `json:"excludeEmpty"`
}

type UserAttrReqData struct {
	UserGroup         []int          `json:"userGroup"`
	ZhibiaoArr        []string       `json:"zhibiaoArr"`
	GroupBy           []string       `json:"groupBy"`
	WhereFilterByUser AnalysisFilter `json:"whereFilterByUser"`
	Appid             int            `json:"appid"`
}

type UserListReqData struct {
	UI    []string `json:"ui"`
	Appid int      `json:"appid"`
}

type NewPannel struct {
	PannelName string `json:"pannel_name"`
	FolderId   int    `json:"folder_id"`
}

type NewDir struct {
	FolderName string `db:"folder_name" json:"folder_name"`
	FolderType int8   `db:"folder_type" json:"folder_type"` //0为自己创建的
	CreateBy   int    `db:"create_by" json:"create_by"`
	Appid      int    `db:"appid" json:"appid"`
}

type FindRtById struct {
	Appid int `db:"appid" json:"appid"`
	Id    int `json:"id"`
}

type FindNameCount struct {
	Appid  int    `db:"appid" json:"appid"`
	Name   string `db:"name" json:"name"`
	RtType int8   `db:"rt_type" json:"rt_type"`
}

type GetPannelList struct {
	Appid int `db:"appid" json:"appid"`
}

type AddUserGroup struct {
	Ids    []string `json:"uids"`
	Name   string   `json:"name"`
	Remark string   `json:"remark"`
	Appid  int      `json:"appid"`
}

type ModifyUserGroup struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Remark string `json:"remark"`
	Appid  int    `json:"appid"`
}

type DeleteUserGroup struct {
	Id    int `json:"id"`
	Appid int `json:"appid"`
}

type UserGroupList struct {
	Appid int `json:"appid"`
}

type UserGroupListReq struct {
	Appid      int    `json:"appid"`
	Keyword    string `json:"keyword"`
	UpdateType string `json:"update_type"`
}

type UserGroupDetailReq struct {
	Id    int `json:"id"`
	Appid int `json:"appid"`
}

type SaveUserGroupReq struct {
	Id               int             `json:"id"`
	Appid            int             `json:"appid"`
	GroupName        string          `json:"group_name"`
	GroupDisplayName string          `json:"group_display_name"`
	Remark           string          `json:"remark"`
	UpdateType       string          `json:"update_type"`
	CreateType       string          `json:"create_type"`
	RuleContent      json.RawMessage `json:"rule_content"`
	SnapshotUserList []string        `json:"snapshot_user_list"`
}

type RefreshUserGroupReq struct {
	Id    int `json:"id"`
	Appid int `json:"appid"`
}

type UserEventDetailReq struct {
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	Appid      int      `json:"appid"`
	UserID     string   `json:"userId"`
	EventName  string   `json:"eventName"`
	OrderBy    string   `json:"orderBy"`
	Date       []string `json:"date"`
	EventNames []string `json:"eventNames"`
}

type UserEventListReq struct {
	Uid   int `json:"uid"`
	Appid int `json:"appid"`
}

type UserEventCountReq struct {
	Appid            int      `json:"appid"`
	WindowTimeFormat string   `json:"windowTimeFormat"`
	UserID           string   `json:"userId"`
	EventNames       []string `json:"eventNames"`
	Date             []string `json:"date"`
}

type LoadPropQuotasReq struct {
	EventName string `json:"event_name"`
	Appid     int    `json:"appid"`
}

type RolesDelReq struct {
	Id int `json:"id"`
}

type GmOperaterLogList struct {
	Page           int      `json:"page"`
	Limit          int      `json:"limit"`
	OperaterAction string   `json:"operater_action"`
	RoleId         int      `json:"role_id"`
	UserId         int      `json:"user_id"`
	Date           []string `json:"date"`
}

type UpdateAttrInvisibleReq struct {
	Appid           int    `json:"appid"`
	AttributeSource int    `json:"attribute_source"`
	AttributeName   string `json:"attribute_name"`
	Status          int    `json:"status"`
}

type AttrManagerByMetaReq struct {
	Appid     int    `json:"appid"`
	Typ       int    `json:"typ"`
	EventName string `json:"event_name"`
}

type UpdateAttrShowNameReq struct {
	Appid         int    `json:"appid"`
	AttributeName string `json:"attribute_name"`
	Typ           int    `json:"typ"`
	ShowName      string `json:"show_name"`
}

type UpdateShowNameReq struct {
	Appid     int    `json:"appid"`
	EventName string `json:"event_name"`
	ShowName  string `json:"show_name"`
}

type GetCalcuSymbolDataReq struct {
	Appid     int    `json:"appid"`
	EventName string `json:"event_name"`
}

type DeleteUserReq struct {
	Id int32 `json:"id"`
}

type GetUserByIdReq struct {
	Id int32 `json:"id"`
}

type UserUpdateReq struct {
	Id       int32  `json:"id"`
	Realname string `json:"realname"`
	RoleId   int32  `json:"role_id"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type UserAddReq struct {
	Realname string `json:"realname"`
	RoleId   int32  `json:"role_id"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type UserBanReq struct {
	Id  int32 `json:"id"`
	Typ int   `json:"typ"`
}

type AttrManagerReq struct {
	Appid int `json:"appid"`
	Typ   int `json:"typ"`
}

type GetAnalyseSelectOptionsReq struct {
	Appid int `json:"appid"`
}

type ReportCountReq struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Appid     int    `json:"appid"`
}

type EventFailDescReq struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Appid     int    `json:"appid"`
	DataName  string `json:"data_name"`
}

type AddDebugDeviceIDReq struct {
	Appid    int    `json:"appid"`
	Remark   string `json:"remark"`
	DeviceID string `json:"device_id"`
}

type DelDebugDeviceIDReq struct {
	Appid    int    `json:"appid"`
	DeviceID string `json:"device_id"`
}

type DebugDeviceIDListReq struct {
	Appid int `json:"appid"`
}
