
[[define "GetByServiceID"]]
{{define "GetByServiceID"}}
select * from `[[.TableName]]` where `service_id`=:ServiceID and  `deleted_at` is null;
{{end}}
[[end]]

[[tplGetByPrimaryKey .]]
[[tplGetAllByPrimaryKeyList .]]

[[tplPaginateWhere .]]
[[tplPaginateTotal .]]
[[tplPaginate .]]

[[tplInsert .]]
[[tplUpdate .]]
[[tplDel .]]
