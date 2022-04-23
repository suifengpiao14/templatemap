

{{define "ddlService"}}
   CREATE TABLE if not exists `service` (
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '标题',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '介绍',
  `version` varchar(20) NOT NULL DEFAULT '' COMMENT '版本',
  `contact_ids` varchar(100) NOT NULL DEFAULT '' COMMENT '联系人',
  `license` char(255) NOT NULL DEFAULT '' COMMENT '协议',
  `security` char(255) NOT NULL DEFAULT '' COMMENT '鉴权',
  `proxy` char(60) NOT NULL DEFAULT '' COMMENT '代理地址',
  `variables` text NOT NULL COMMENT 'json字符串',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime  DEFAULT null COMMENT '删除时间',
  PRIMARY KEY (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT="项目/服务组件";
{{end}}


{{define "ddlApi"}}
  CREATE TABLE  if not exists `api` (
  `api_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'api 文档标识',
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `name` varchar(255) NOT NULL COMMENT '路由名称(英文)',
  `title` varchar(255) NOT NULL COMMENT '标题',
  `tags` char(255) NOT NULL DEFAULT '' COMMENT '标签ID集合',
  `uri` varchar(255) NOT NULL DEFAULT '' COMMENT '路径',
  `summary` varchar(255) NOT NULL DEFAULT '' COMMENT '摘要',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '介绍',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime  DEFAULT null COMMENT '删除时间',
  PRIMARY KEY (`api_id`),
  KEY `name` (`name`),
  key `server_id` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT="接口";
{{end}}


{{define "ddlParameter"}}
  CREATE TABLE  if not exists `parameter`  (
  `parameter_id` varchar(64) NOT NULL DEFAULT '' COMMENT '参数标识',
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `api_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'api 文档标识',
  `schema_id` varchar(64) NOT NULL DEFAULT '' COMMENT '验证规则标识',
  `full_name` varchar(255) NOT NULL DEFAULT '' COMMENT '全称',
  `name` char(50) NOT NULL DEFAULT '' COMMENT '名称(冗余local.en)',
  `title` char(50) NOT NULL DEFAULT '' COMMENT '名称(冗余local.zh)',
  `type` enum('string','int','number','array','object') NOT NULL DEFAULT 'string' COMMENT '参数类型(string-字符,int-整型,number-数字,array-数组,object-对象)',
  `tag` char(50) NOT NULL DEFAULT '' COMMENT '所属标签',
  `method` enum("get","post","put","delete","head")  NOT NULL DEFAULT 'POST' COMMENT '请求方法(get-GET,post-POST,put-PUT,delete-DELETE,head-HEAD)',
  `http_status` char(10)  NOT NULL DEFAULT '' COMMENT 'http状态码',
  `position` enum('body','head','path','query','cookie') DEFAULT 'body' COMMENT '参数所在的位置(body-BODY,head-HEAD,path-PATH,query-QUERY,cookie-COOKIE)',
  `example` varchar(1024) NOT NULL DEFAULT '' COMMENT '案例',
  `deprecated` enum("true","false")  NOT NULL DEFAULT 'false' COMMENT '是否弃用(true-是,false-否)',
  `required` enum('true','false') NOT NULL DEFAULT 'false' COMMENT '是否必须(true-是,false-否)',
  `serialize` varchar(50) NOT NULL DEFAULT '' COMMENT '对数组、对象序列化方法,参照openapi parameters.style',
  `explode`  enum('true','false') NOT NULL DEFAULT 'true' COMMENT '对象的key,是否单独成参数方式,参照openapi parameters.explode(true-是,false-否)',
  `allow_empty_value` enum('true','false') NOT NULL DEFAULT 'true' COMMENT '是否容许空值(true-是,false-否)',
  `allow_reserved` enum('true','false') NOT NULL DEFAULT 'false' COMMENT '特殊字符是否容许出现在uri参数中(true-是,false-否)',
  `description` varchar(255) NOT NULL COMMENT '简介',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime DEFAULT null COMMENT '删除时间',
  PRIMARY KEY (`parameter_id`),
  UNIQUE KEY `unique_key` (`service_id`,`api_id`,`full_name`,`deleted_at`) USING BTREE,
  key `server_id` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='请求/响应参数';
{{end}}


{{define "ddlSchema"}}
  CREATE TABLE  if not exists `schema` (
  `schema_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'api schema 标识',
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '所属服务标识',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '描述',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `type` enum('integer','array','string','object') NOT NULL COMMENT '类型(integer-整数,array-数组,string-字符串,object-对象)',
  `example` varchar(1024) NOT NULL DEFAULT '' COMMENT '案例',
  `deprecated` enum("true","false")  NOT NULL DEFAULT 'false' COMMENT '是否弃用(true-是,false-否)',
  `required` enum('true','false') NOT NULL DEFAULT 'false' COMMENT '是否必须(true-是,false-否)',
  `enum` char(100) NOT NULL DEFAULT '' COMMENT '枚举值',
  `enum_names` char(100) NOT NULL DEFAULT '' COMMENT '枚举名称',
  `enum_titles` char(100) NOT NULL DEFAULT '' COMMENT '枚举标题',
  `format` char(100) NOT NULL DEFAULT '' COMMENT '格式',
  `default` varchar(255) NOT NULL DEFAULT '' COMMENT '默认值',
  `nullable`  enum('true','false') NOT NULL DEFAULT 'false' COMMENT '是否可以为空(true-是,false-否)',
  `multiple_of` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '倍数',
  `maxnum` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最大值',
  `exclusive_maximum` enum('true','false') NOT NULL DEFAULT 'false' NOT NULL DEFAULT 'false' COMMENT '是否不包含最大项(true-是,false-否)',
  `minimum` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最小值',
  `exclusive_minimum` enum('true','false') NOT NULL DEFAULT 'false' NOT NULL DEFAULT 'true' COMMENT '是否不包含最小项(true-是,false-否)',
  `max_length` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最大长度',
  `min_length` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最小长度',
  `pattern` varchar(255) NOT NULL DEFAULT '' COMMENT '正则表达式',
  `max_items` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最大项数',
  `min_items` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最小项数',
  `unique_items`enum('true','false') NOT NULL DEFAULT 'false' COMMENT '所有项是否需要唯一(true-是,false-否)',
  `max_properties` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最多属性项',
  `min_properties` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最少属性项',
  `all_of` varchar(255) NOT NULL DEFAULT '' COMMENT '所有',
  `one_of` varchar(255) NOT NULL DEFAULT '' COMMENT '只满足一个',
  `any_of` varchar(255) NOT NULL DEFAULT '' COMMENT '任何一个SchemaID',
  `allow_empty_value` enum('true','false') NOT NULL DEFAULT 'true' COMMENT '是否容许空值(true-是,false-否)',
  `allow_reserved` enum('true','false') NOT NULL DEFAULT 'false' COMMENT '特殊字符是否容许出现在uri参数中(true-是,false-否)',
  `not` varchar(255) NOT NULL DEFAULT '' COMMENT '不包含的schemaID',
  `additional_properties` varchar(255) NOT NULL DEFAULT '' COMMENT 'boolean',
  `discriminator` varchar(255) NOT NULL DEFAULT '' COMMENT 'schema鉴别',
  `read_only` enum('true','false') NOT NULL DEFAULT 'false' COMMENT '是否只读(true-是,false-否)',
  `write_only` enum('true','false') NOT NULL DEFAULT 'false' COMMENT '是否只写(true-是,false-否)',
  `xml` varchar(255) NOT NULL DEFAULT '' COMMENT 'xml对象',
  `external_docs` varchar(255)  NOT NULL DEFAULT '' COMMENT '附加文档',
  `external_pros` varchar(255)  NOT NULL DEFAULT '' COMMENT '附加文档',
  `extensions` varchar(255) NOT NULL DEFAULT '' COMMENT '扩展字段',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime  DEFAULT null COMMENT '删除时间',
  `summary` varchar(255) NOT NULL DEFAULT '' COMMENT '简介',
  PRIMARY KEY (`schema_id`),
  key `server_id` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='响应格式';
{{end}}



{{define "ddlExample"}}
CREATE TABLE  if not exists `example`  (
  `example_id` varchar(64) NOT NULL DEFAULT '' COMMENT '测试用例标识',
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `api_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'api 文档标识',
  `tag` varchar(50) NOT NULL DEFAULT '' COMMENT '标签,mock数据时不同接口案例优先返回相同tag案例',
  `title` varchar(60) NOT NULL DEFAULT '' COMMENT '案例名称',
  `summary` varchar(255) NOT NULL COMMENT '简介',
  `url` varchar(255) NOT NULL COMMENT 'URL',
  `method` enum("get","post","put","delete","head")  NOT NULL DEFAULT 'POST' COMMENT '请求方法(get-GET,post-POST,put-PUT,delete-DELETE,head-HEAD)',
  `pre_request_script` text NOT NULL default "" comment '前置请求脚本',
  `auth` text NOT NULL default "" comment 'query 鉴权',
  `headers` text NOT NULL default "" comment 'query 请求头',
  `parameters` text NOT NULL default "" comment 'query 请求参数(json序列化)',
  `content_type` enum("application/json","plain/text") NOT NULL default "application/json" comment '请求格式(application/json-json,plain/text-文本)',
  `body` text NOT NULL default "" comment 'query 请求体',
  `test_script` text NOT NULL default "" comment 'query 返回体测试脚本',
  `response` text NOT NULL default "" comment '请求体',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime  DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`example_id`),
  key `server_id` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='案例';
{{end}}

{{define "ddlMarkdown"}}
CREATE TABLE  if not exists `markdown`  (
  `markdown_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'markdown 文档标识',
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `api_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'api 文档标识',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '唯一名称',
  `title` varchar(100) NOT NULL DEFAULT '' COMMENT '文章标题',
  `markdown` longtext,
  `content` longtext,
  `owner_id` bigint(11) NOT NULL DEFAULT '0' COMMENT '作者ID',
  `owner_name` varchar(100) NOT NULL DEFAULT '' COMMENT '作者名称',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime  DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`markdown_id`),
  UNIQUE KEY `name` (`name`),
  key `server_id` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='markdown 文档';
{{end}}

{{define "ddlServer"}}
  CREATE TABLE  if not exists `server`  (
  `server_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `service_id` varchar(64) NOT NULL DEFAULT '' COMMENT '服务标识',
  `url` varchar(255) NOT NULL DEFAULT '' COMMENT '服务器地址',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '介绍',
  `proxy` varchar(30) NOT NULL DEFAULT '' COMMENT '代理地址',
  `extension_ids` varchar(255) NOT NULL DEFAULT '' COMMENT '扩展字段',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `deleted_at` datetime  DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`server_id`),
  key `server_id` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='服务器部署';
{{end}}






