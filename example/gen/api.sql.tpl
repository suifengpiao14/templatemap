

{{define "GetAllByAPIIDList"}}
select * from `api` where `api_id` in ({{in . .APIIDList}})  and `deleted_at` is null limit :Offset,:Limit;
{{end}}

{{define "PaginateWhere"}}
{{end}}

{{define "PaginateTotal"}}
select count(*) as `count` from `api` where 1=1 {{template "PaginateWhere" .}} and `deleted_at` is null;
{{setValue . "Offset" (mul .PageIndex  .PageSize)}}
{{setValue . "Limit" .PageSize}}
{{ $list:=executeTemplate . "GetAllByAPIIDList"| toSQL . | execSQL . "db_identifier"}}
{{if $list}}
{{setValue . "GetAllByAPIIDList" $list}}
{{$list}}
{{end}}

{{end}}
