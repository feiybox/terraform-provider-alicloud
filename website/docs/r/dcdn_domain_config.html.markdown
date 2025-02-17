---
subcategory: "DCDN"
layout: "alicloud"
page_title: "Alicloud: alicloud_dcdn_domain_config"
sidebar_current: "docs-alicloud-resource-dcdn-domain-config"
description: |-
  Provides a Alicloud Dcdn domain config Resource.
---

# alicloud_dcdn_domain_config

Provides a DCDN Accelerated Domain resource.

For information about domain config and how to use it, see [Batch set config](https://www.alibabacloud.com/help/en/doc-detail/130632.htm)

-> **NOTE:** Available since v1.131.0.

## Example Usage

Basic Usage

```terraform
variable "domain_name" {
  default = "example.com"
}
resource "alicloud_dcdn_domain" "example" {
  domain_name = var.domain_name
  scope       = "overseas"
  sources {
    content  = "1.1.1.1"
    port     = "80"
    priority = "20"
    type     = "ipaddr"
    weight   = "10"
  }
}

resource "alicloud_dcdn_domain_config" "example" {
  domain_name   = alicloud_dcdn_domain.example.domain_name
  function_name = "ip_allow_list_set"
  function_args {
    arg_name  = "ip_list"
    arg_value = "110.110.110.110"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required, ForceNew) Name of the accelerated domain. This name without suffix can have a string of 1 to 63 characters, must contain only alphanumeric characters or "-", and must not begin or end with "-", and "-" must not in the 3th and 4th character positions at the same time. Suffix `.sh` and `.tel` are not supported.
* `function_name` - (Required, ForceNew) The name of the domain config.
* `function_args` - (Required, List) The args of the domain config.  See [`function_args`](#function_args) below.

### `function_args`

The `function_args` block supports the following:

* `arg_name` - (Required) The name of arg.
* `arg_value` - (Required) The value of arg.

## Attributes Reference

The following attributes are exported:

* `config_id` - The DCDN domain config id.
* `id` - The ID of this resource. The value is formate as `<domain_name>:<function_name>:<config_id>`.
* `status` -  The status of this resource.

## Import

DCDN domain config can be imported using the id, e.g.

```shell
$ terraform import alicloud_dcdn_domain_config.example <domain_name>:<function_name>:<config_id>
```
