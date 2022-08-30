package alicloud

import (
	"fmt"
	"log"
	"time"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudNasSmbAclAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudNasSmbAclAttachmentCreate,
		Read:   resourceAlicloudNasSmbAclAttachmentRead,
		Update: resourceAlicloudNasSmbAclAttachmentUpdate,
		Delete: resourceAlicloudNasSmbAclAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"keytab": {
				Type:     schema.TypeString,
				Required: true,
			},
			"keytab_md5": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enable_anonymous_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"encrypt_data": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"reject_unencrypted_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"super_admin_sid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"home_dir_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 32767),
			},
			"enabled": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAlicloudNasSmbAclAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "EnableSmbAcl"
	request := make(map[string]interface{})
	conn, err := client.NewNasClient()
	if err != nil {
		return WrapError(err)
	}
	request["RegionId"] = client.Region
	request["FileSystemId"] = d.Get("file_system_id")
	request["Keytab"] = d.Get("keytab")
	request["KeytabMd5"] = d.Get("keytab_md5")

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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_smb_acl_attachment", action, AlibabaCloudSdkGoERROR)
	}
	d.SetId(fmt.Sprint(request["FileSystemId"], ":", request["Keytab"], ":", request["KeytabMd5"]))
	return resourceAlicloudNasSmbAclAttachmentRead(d, meta)
}

func resourceAlicloudNasSmbAclAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
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
		"RegionId":     client.RegionId,
		"FileSystemId": parts[0],
	}

	update := false

	if d.HasChange("keytab") {
		update = true
	}
	request["Keytab"] = d.Get("keytab")

	if d.HasChange("keytab_md5") {
		update = true
	}
	request["KeytabMd5"] = d.Get("keytab_md5")

	if d.HasChange("enable_anonymous_access") {
		update = true
	}
	request["EnableAnonymousAccess"] = d.Get("enable_anonymous_access")

	if d.HasChange("encrypt_data") {
		update = true
	}
	request["EncryptData"] = d.Get("encrypt_data")

	if d.HasChange("reject_unencrypted_access") {
		update = true
	}
	request["RejectUnencryptedAccess"] = d.Get("reject_unencrypted_access")

	if d.HasChange("super_admin_sid") {
		update = true
	}
	request["SuperAdminSid"] = d.Get("super_admin_sid")

	if d.HasChange("home_dir_path") {
		update = true
	}
	request["HomeDirPath"] = d.Get("home_dir_path")

	if update {
		action := "ModifySmbAcl"
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
	return resourceAlicloudNasSmbAclAttachmentRead(d, meta)
}

func resourceAlicloudNasSmbAclAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService := NasService{client}
	object, err := nasService.DescribeNasSmbAcl(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_smb_acl_attachment nasService.DescribeNasSmbAcl Failed!!! %s",
				err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	d.Set("file_system_id", parts[0])
	d.Set("keytab", parts[1])
	d.Set("keytab_md5", parts[2])
	d.Set("auth_method", fmt.Sprint(object["AuthMethod"]))
	d.Set("enable_anonymous_access", fmt.Sprint(object["EnableAnonymousAccess"]))
	d.Set("encrypt_data", fmt.Sprint(object["EncryptData"]))
	d.Set("reject_unencrypted_access", fmt.Sprint(object["RejectUnencryptedAccess"]))
	d.Set("super_admin_sid", fmt.Sprint(object["SuperAdminSid"]))
	d.Set("home_dir_path", fmt.Sprint(object["HomeDirPath"]))
	d.Set("enabled", fmt.Sprint(object["Enabled"]))
	return nil
}

func resourceAlicloudNasSmbAclAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DisableSmbAcl"
	var response map[string]interface{}
	conn, err := client.NewNasClient()
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	request := map[string]interface{}{
		"RegionId":     client.RegionId,
		"FileSystemId": parts[0],
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
