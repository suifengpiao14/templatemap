# templatemap
### 计划
将 validate_schema、output_schema 改成 key:value 格式(类似go tag)，提升可阅读性和维护效率
出参案例：
```
{"$schema":"http://json-schema.org/draft-07/schema#","$id":"execAPIXyxzBlacklistInfoPaginate","properties":{"items":{"type":"array","items":{"type":"object","properties":{"id":{"type":"string","src":"PaginateOut.#.Fid"},"openId":{"type":"string","src":"PaginateOut.#.Fopen_id"},"type":{"type":"string","src":"PaginateOut.#.Fopen_id_type"},"status":{"type":"string","src":"PaginateOut.#.Fstatus"}},"required":["id","openId","type","status"]}},"pageInfo":{"type":"object","properties":{"pageIndex":{"type":"string","src":"input.pageIndex"},"pageSize":{"type":"string","src":"input.pageSize"},"total":{"type":"string","src":"PaginateTotalOut"}},"required":["pageIndex","pageSize","total"]}},"type":"object","required":["items","pageInfo"]}
```
转换为：
```
name:items[].items.id,src:PaginateOut.#.Fid,type:string,required:true
name:items[].items.openId,src:PaginateOut.#.Fopen_id,type:string,required:true
name:items[].items.type,src:PaginateOut.#.Fopen_id_type,type:string,required:true
name:items[].items.status,src:PaginateOut.#.Fstatus,type:string,required:true
name:pageInfo.pageIndex,src:input.pageIndex,type:string,required:true
name:pageInfo.pageSize,src:input.pageSize,type:string,required:true
name:pageInfo.total,src:PaginateTotalOut,type:string,required:true
```

入参案例：
```
{"$schema":"http://json-schema.org/draft-07/schema#","$id":"execAPIXyxzBlacklistInfoInsert","properties":{"config":{"type":"object","properties":{"openId":{"type":"string","format":"DBValidate"},"type":{"type":"string","format":"number","enum":["1","2"]},"status":{"type":"string","format":"number","enum":["0","1"]}},"required":["openId","type","status"]}},"type":"object"}
```
转换为：
```
name:config.openId,dst:FopenID,format:DBValidate,type:string,required:true
name:config.type,dst:FopenIDType,enum:["1","2"],type:string,required:true
name:config.status,dst:Fstatus,enum:["0","1"],type:string,required:true

```
分页入参案例:
```
{"$schema":"http://json-schema.org/draft-07/schema#","$id":"execAPIInquiryScreenIdentifyPaginate","properties":{"pageSize":{"type":"string","format":"number"},"pageIndex":{"type":"string","format":"number"}},"type":"object","required":["pageSize","pageIndex"]}
```
转换为:
```
name:pageSize,dst:limit,format:number,type:string,required:true
name:pageIndex,dst:Offset,format:number,type:string,required:true,tpl:{{setValue . "Offset" (mul  (getValue .  "input.pageIndex")   (getValue . "input.pageSize"))}}
```
dst、tpl 字段会提炼出，动态生成 template内容
通过这种转换后，更易于书写和阅读，和计划中的文档格式更相似，同时dst、tpl等字段定义优化了值转换的管理为自动校验提供可行的机制，后续api 的 exec 字段可能被弃用
