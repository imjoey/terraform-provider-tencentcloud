---
layout: "tencentcloud"
page_title: "TencentCloud: tencentcloud_mysql_account_privilege"
sidebar_current: "docs-tencentcloud-resource-mysql_account_privilege"
description: |-
 Provides a mysql account privilege resource to grant different access privilege to different database. A database can be granted by multiple account.
---

# tencentcloud_mysql_account_privilege

Provides a mysql account privilege resource to grant different access privilege to different database. A database can be granted by multiple account.

## Example Usage

```hcl
resource "tencentcloud_mysql_account_privilege" "default" {
  mysql_id = "my-test-database"
  account_name= "tf_account"
  privileges = ["SELECT"]
  database_names = ["instance.name"]
}
```

## Argument Reference

The following arguments are supported:

* `account_name` - (Required, ForceNew) Account name.
* `database_names` - (Required) List of specified database name.
* `mysql_id` - (Required, ForceNew) Instance ID.
* `privileges` - (Optional) Database permissions. Available values for Privileges: "SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "REFERENCES", "INDEX", "ALTER", "CREATE TEMPORARY TABLES", "LOCK TABLES","EXECUTE", "CREATE VIEW", "SHOW VIEW", "CREATE ROUTINE", "ALTER ROUTINE", "EVENT", and "TRIGGER".

