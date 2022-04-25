

{{define "GetAllByAPIIDList"}}
select * from `api` where `api_id` in ({{in . .APIIDList}})  and `deleted_at` is null limit :Offset,:Limit;
{{end}}

{{define "PaginateWhere"}}
{{end}}


{{define "PaginateTotal"}}
select count(*) as `count` from `api` where 1=1 {{template "PaginateWhere" .}} and `deleted_at` is null;
{{end}}




{{define "Paginate"}}
select * from `api` where 1=1 {{template "PaginateWhere" .}} and `deleted_at` is null limit :Offset,:Limit ;
{{end}}



{{define "getPaginate"}}
{{$total:=executeTemplate . "PaginateTotal"|toSQL . | exec . "docapi_db2"}}
{{setValue . "pagination.total" $total}}
{{if $total}}
    {{setValue . "Offset" (mul .PageIndex  .PageSize)}}
    {{setValue . "Limit" (atoi .PageSize)}}
    {{executeTemplate . "Paginate"|toSQL . | exec . "docapi_db2"|setValue . "items"}}
{{end}}
   

{{end}}

