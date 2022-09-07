package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudNasAccessRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudNasAccessRuleCreate,
		Read:   resourceAlicloudNasAccessRuleRead,
		Update: resourceAlicloudNasAccessRuleUpdate,
		Delete: resourceAlicloudNasAccessRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"access_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_cidr_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rw_access_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"RDWR", "RDONLY"}, false),
				Default:      "RDWR",
			},
			"user_access_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"no_squash", "root_squash", "all_squash"}, false),
				Default:      "no_squash",
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 100),
			},
			"access_rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"extreme", "standard"}, false),
				Default:      "standard",
			},
			"ipv6_source_cidr_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"source_cidr_ip"},
			},
		},
	}
}

func resourceAlicloudNasAccessRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateAccessRule"
	request := make(map[string]interface{})
	conn, err := client.NewNasClient()
	if err != nil {
		return WrapError(err)
	}
	request["RegionId"] = client.Region
	request["AccessGroupName"] = d.Get("access_group_name")

	if v, ok := d.GetOk("rw_access_type"); ok && v.(string) != "" {
		request["RWAccessType"] = v
	}
	if v, ok := d.GetOk("user_access_type"); ok && v.(string) != "" {
		request["UserAccessType"] = v
	}
	if v, ok := d.GetOk("source_cidr_ip"); ok && v.(string) != "" {
		request["SourceCidrIp"] = d.Get("source_cidr_ip")
	}
	if v, ok := d.GetOk("ipv6_source_cidr_ip"); ok && v.(string) != "" {
		request["Ipv6SourceCidrIp"] = d.Get("ipv6_source_cidr_ip")
	}

	request["Priority"] = d.Get("priority")
	request["FileSystemType"] = d.Get("file_system_type")

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2017-06-26"), StringPointer("AK"), nil, request, &util.RuntimeOptions{})
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_rule", action, AlibabaCloudSdkGoERROR)
	}
	d.SetId(fmt.Sprint(request["AccessGroupName"], ":", response["AccessRuleId"], ":", request["FileSystemType"]))
	return resourceAlicloudNasAccessRuleRead(d, meta)
}

func resourceAlicloudNasAccessRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	conn, err := client.NewNasClient()
	if err != nil {
		return WrapError(err)
	}
	var response map[string]interface{}
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		err = WrapError(err)
		return err
	}
	request := map[string]interface{}{
		"RegionId":        client.RegionId,
		"AccessGroupName": parts[0],
		"AccessRuleId":    parts[1],
		"FileSystemType":  parts[2],
	}

	update := false

	if d.HasChange("rw_access_type") {
		update = true
	}
	request["RWAccessType"] = d.Get("rw_access_type")

	if d.HasChange("user_access_type") {
		update = true
	}
	request["UserAccessType"] = d.Get("user_access_type")

	if d.HasChange("priority") {
		update = true
	}
	request["Priority"] = d.Get("priority")

	if v, ok := d.GetOk("ipv6_source_cidr_ip"); ok && v.(string) != "" {
		update = true
		request["Ipv6SourceCidrIp"] = d.Get("ipv6_source_cidr_ip")
	}
	if v, ok := d.GetOk("source_cidr_ip"); ok && v.(string) != "" {
		update = true
		request["SourceCidrIp"] = d.Get("source_cidr_ip")
	}

	if update {
		action := "ModifyAccessRule"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2017-06-26"), StringPointer("AK"), nil, request, &util.RuntimeOptions{})
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			addDebug(action, response, request)
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
	}
	return resourceAlicloudNasAccessRuleRead(d, meta)
}

func resourceAlicloudNasAccessRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService := NasService{client}
	object, err := nasService.DescribeNasAccessRule(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_access_rule nasService.DescribeNasAccessRule Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), ":")
	d.Set("access_group_name", parts[0])
	d.Set("access_rule_id", parts[1])
	if len(parts) == 2 {
		d.SetId(fmt.Sprintf("%s:%s:%s", parts[0], parts[1], "standard"))
		d.Set("file_system_type", "standard")
	} else {
		d.Set("file_system_type", parts[2])
	}

	d.Set("source_cidr_ip", object["SourceCidrIp"])
	d.Set("priority", formatInt(object["Priority"]))
	d.Set("rw_access_type", object["RWAccess"])
	d.Set("user_access_type", object["UserAccess"])
	d.Set("ipv6_source_cidr_ip", object["Ipv6SourceCidrIp"])

	return nil
}

func resourceAlicloudNasAccessRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteAccessRule"
	var response map[string]interface{}
	conn, err := client.NewNasClient()
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		err = WrapError(err)
		return err
	}
	request := map[string]interface{}{
		"RegionId":        client.RegionId,
		"AccessGroupName": parts[0],
		"AccessRuleId":    parts[1],
		"FileSystemType":  parts[2],
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = conn.DoRequest(StringPointer(action), nil, StringPointer("POST"), StringPointer("2017-06-26"), StringPointer("AK"), nil, request, &util.RuntimeOptions{})
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		if IsExpectedErrors(err, []string{"Forbidden.NasNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}
	return nil
}
