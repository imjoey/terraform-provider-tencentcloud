package tencentcloud

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	tencentCloudApiInstanceChargeTypePrePaid        = "PREPAID"
	tencentCloudApiInstanceChargeTypePostPaidByHour = "POSTPAID_BY_HOUR"
)

const (
	tencentCloudApiInternetChargeTypeBandwithPrepaid         = "BANDWIDTH_PREPAID"
	tencentCloudApiInternetChargeTypeTrafficPostpaidByHour   = "TRAFFIC_POSTPAID_BY_HOUR"
	tencentCloudApiInternetChargeTypeBandwidthPostpaidByHour = "BANDWIDTH_POSTPAID_BY_HOUR"
	tencentCloudApiInternetChargeTypeBandwidthPackage        = "BANDWIDTH_PACKAGE"
)

const (
	tencentCloudApiInstanceChargeTypePrePaidRenewFlagNotifyAndAutoRenew          = "NOTIFY_AND_AUTO_RENEW"
	tencentCloudApiInstanceChargeTypePrePaidRenewFlagNotifyAndManualRenew        = "NOTIFY_AND_MANUAL_RENEW"
	tencentCloudApiInstanceChargeTypePrePaidRenewFlagDisableNotifyAndManualRenew = "DISABLE_NOTIFY_AND_MANUAL_RENEW"
)

const (
	tencentCloudApiDiskTypeLocalBaisc   = "LOCAL_BASIC"
	tencentCloudApiDiskTypeLocalSSD     = "LOCAL_SSD"
	tencentCloudApiDiskTypeCloudBasic   = "CLOUD_BASIC"
	tencentCloudApiDiskTypeCloudSSD     = "CLOUD_SSD"
	tencentCloudApiDiskTypeCloudPremium = "CLOUD_PREMIUM"
)

var (
	availableInstanceChargeTypes = []string{
		tencentCloudApiInstanceChargeTypePrePaid,
		tencentCloudApiInstanceChargeTypePostPaidByHour,
	}
	availableInternetChargeTypes = []string{
		tencentCloudApiInternetChargeTypeBandwithPrepaid,
		tencentCloudApiInternetChargeTypeTrafficPostpaidByHour,
		tencentCloudApiInternetChargeTypeBandwidthPostpaidByHour,
		tencentCloudApiInternetChargeTypeBandwidthPackage,
	}
	availableDiskTypes = []string{
		tencentCloudApiDiskTypeLocalBaisc,
		tencentCloudApiDiskTypeLocalSSD,
		tencentCloudApiDiskTypeCloudBasic,
		tencentCloudApiDiskTypeCloudSSD,
		tencentCloudApiDiskTypeCloudPremium,
	}
	availableInstanceChargeTypePrePaidPeriodValues    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36}
	availableInstanceChargeTypePrePaidRenewFlagValues = []string{
		tencentCloudApiInstanceChargeTypePrePaidRenewFlagNotifyAndAutoRenew,
		tencentCloudApiInstanceChargeTypePrePaidRenewFlagNotifyAndManualRenew,
		tencentCloudApiInstanceChargeTypePrePaidRenewFlagDisableNotifyAndManualRenew,
	}
)

var (
	// TODO remove me when related feature implemented
	unsupportedUpdateFields = []string{
		"instance_charge_type_prepaid_period",
		"instance_charge_type_prepaid_renew_flag",
		"internet_charge_type",
		"internet_max_bandwidth_out",
		"allocate_public_ip",
		"system_disk_size",
		"data_disks",
		// we can remove it once tag api support default uin when it is absent in resource URI
		"tags",
	}
)

func resourceTencentCloudInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudInstanceCreate,
		Read:   resourceTencentCloudInstanceRead,
		Update: resourceTencentCloudInstanceUpdate,
		Delete: resourceTencentCloudInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terrafrom-CVM-Instance",
				ValidateFunc: validateInstanceName,
			},
			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateInstanceType,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
			},
			// payment
			"instance_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateInstanceChargeType,
			},
			"instance_charge_type_prepaid_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateInstanceChargeTypePrePaidPeriod,
			},
			"instance_charge_type_prepaid_renew_flag": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateInstanceChargeTypePrePaidRenewFlag,
			},
			// network
			"internet_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateInternetChargeType,
			},
			"internet_max_bandwidth_out": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateInternetMaxBandwidthOut,
			},
			"allocate_public_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			// vpc
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			// security group
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			// storage
			"system_disk_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateDiskType,
			},
			// TODO finish me when integrating CBS
			//"system_disk_id": &schema.Schema{
			//	Type:     schema.TypeString,
			//	Optional: true,
			//},
			"system_disk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateIntegerInRange(50, 1000),
			},
			"data_disks": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_disk_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateDiskType,
						},
						// TODO finish me when integrating CBS
						//"data_disk_id": &schema.Schema{
						//	Type:     schema.TypeString,
						//	Optional: true,
						//},
						"data_disk_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validateDiskSize,
						},
						"delete_with_instance": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			// enhance services
			"disable_security_service": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"disable_monitor_service": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// login
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user_data_raw"},
			},
			"user_data_raw": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"user_data"},
			},
			// cvm api 2017-03-12 runinstances defines tags as a list of map,
			// such as:
			//"tags": {
			//	Type:     schema.TypeList,
			//	Optional: true,
			//	Elem: schema.Resource{
			//		Schema: map[string]*schema.Schema{
			//			"resource_type": {
			//				Type:     schema.TypeString,
			//				Required: true,
			//			},
			//			"tags": {
			//				Type:     schema.TypeList,
			//				Required: true,
			//				Elem: schema.Resource{
			//					Schema: map[String]*schema.Schema{
			//						"key": {
			//							Type:     schema.TypeString,
			//							Required: true,
			//						},
			//						"value": {
			//							Type:     schema.TypeString,
			//							Required: true,
			//						},
			//					},
			//				},
			//			},
			//		},
			//	},
			//},
			// but it actually only accept "instance" as resource type, and list will be merged into
			// key:value pairs, which means it can be presented as following:
			// Note that aws has exactly same definition in API spec and same schema in terraform,
			// here we follow them.
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
			},

			// Computed values.
			"instance_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTencentCloudInstanceCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*TencentCloudClient).commonConn

	params := map[string]string{
		"Version":        "2017-03-12",
		"Action":         "RunInstances",
		"Placement.Zone": d.Get("availability_zone").(string),
		"ImageId":        d.Get("image_id").(string),
	}

	if instanceType, ok := d.GetOk("instance_type"); ok {
		insType := instanceType.(string)
		if len(insType) > 0 {
			params["InstanceType"] = insType
		}
	}
	if instanceName, ok := d.GetOk("instance_name"); ok {
		insName := instanceName.(string)
		if len(insName) > 0 {
			params["InstanceName"] = insName
		}
	}
	if hostName, ok := d.GetOk("hostname"); ok {
		hName := hostName.(string)
		if len(hName) > 0 {
			params["HostName"] = hName
		}
	}
	if projectId, ok := d.GetOk("project_id"); ok {
		params["Placement.ProjectId"] = fmt.Sprintf("%v", projectId.(int))
	}

	if instanceChargeType, ok := d.GetOk("instance_charge_type"); ok {
		insChargeType := instanceChargeType.(string)
		if insChargeType == tencentCloudApiInstanceChargeTypePrePaid {
			period, ok := d.GetOk("instance_charge_type_prepaid_period")
			if !ok {
				return fmt.Errorf(
					"tencentcloud_instance instance_charge_type_prepaid_period is need when instance_charge_type is %v",
					tencentCloudApiInstanceChargeTypePrePaid,
				)
			}
			periodStr := fmt.Sprintf("%v", period.(int))
			params["InstanceChargePrepaid.Period"] = periodStr
			if renewFlag, ok := d.GetOk("instance_charge_type_prepaid_renew_flag"); ok {
				params["InstanceChargePrepaid.RenewFlag"] = renewFlag.(string)
			}
		}

		params["InstanceChargeType"] = insChargeType
	}

	// network releated
	if internetChargeType, ok := d.GetOk("internet_charge_type"); ok {
		netChargeType := internetChargeType.(string)
		params["InternetAccessible.InternetChargeType"] = netChargeType
	}
	if internetMaxBandwidthOut, ok := d.GetOk("internet_max_bandwidth_out"); ok {
		maxBandwidthOut := internetMaxBandwidthOut.(int)
		params["InternetAccessible.InternetMaxBandwidthOut"] = fmt.Sprintf("%v", maxBandwidthOut)
	}
	// assign public IP
	if allocatePublicIP, ok := d.Get("allocate_public_ip").(bool); ok {
		if allocatePublicIP {
			params["InternetAccessible.PublicIpAssigned"] = "TRUE"
		} else {
			params["InternetAccessible.PublicIpAssigned"] = "FALSE"
		}
	}

	// security groups
	if v, ok := d.GetOk("security_groups"); ok {
		sgIds := v.(*schema.Set).List()
		if len(sgIds) > 0 {
			for i, sgId := range sgIds {
				paramKey := fmt.Sprintf("SecurityGroupIds.%v", i)
				params[paramKey] = sgId.(string)
			}
		}
	}

	// storage
	if systemDiskType, ok := d.GetOk("system_disk_type"); ok {
		params["SystemDisk.DiskType"] = systemDiskType.(string)
	}
	if systemDiskSize, ok := d.GetOk("system_disk_size"); ok {
		diskSize := systemDiskSize.(int)
		params["SystemDisk.DiskSize"] = fmt.Sprintf("%v", diskSize)
	}
	var dataDisksAttr []map[string]interface{}
	if dataDisks, ok := d.GetOk("data_disks"); ok {
		dataDiskList := dataDisks.([]interface{})
		if len(dataDiskList) > 10 {
			return fmt.Errorf("Too many data disks for tencentcloud_instance!")
		}
		for i, dataDisk := range dataDiskList {
			dd := dataDisk.(map[string]interface{})
			if v, ok := dd["data_disk_type"].(string); ok && v != "" {
				paramKey := fmt.Sprintf("DataDisks.%v.DiskType", i)
				params[paramKey] = v
			}
			if v, ok := dd["data_disk_size"].(int); ok {
				paramKey := fmt.Sprintf("DataDisks.%v.DiskSize", i)
				paramValue := fmt.Sprintf("%v", v)
				params[paramKey] = paramValue
			}
			if v, ok := dd["delete_with_instance"].(bool); ok {
				paramKey := fmt.Sprintf("DataDisks.%v.DeleteWithInstance", i)
				if v {
					params[paramKey] = "TRUE"
				} else {
					params[paramKey] = "FALSE"
				}
			}

			dataDisksAttr = append(dataDisksAttr, dd)
		}
	}

	// enhance services
	if v, ok := d.GetOk("disable_security_service"); ok {
		disable := v.(bool)
		if disable {
			params["EnhancedService.SecurityService.Enabled"] = "FALSE"
		}
	}
	if v, ok := d.GetOk("disable_monitor_service"); ok {
		disable := v.(bool)
		if disable {
			params["EnhancedService.MonitorService.Enabled"] = "FALSE"
		}
	}

	// login confidential
	if v, ok := d.GetOk("key_name"); ok {
		keyId := v.(string)
		params["LoginSettings.KeyIds.0"] = keyId
	}
	if v, ok := d.GetOk("password"); ok {
		passwd := v.(string)
		params["LoginSettings.Password"] = passwd
	}
	if v, ok := d.GetOk("user_data"); ok {
		data := v.(string)
		if len(data) > 0 {
			params["UserData"] = data
		}
	}
	if v, ok := d.GetOk("user_data_raw"); ok {
		data := v.(string)
		if len(data) > 0 {
			params["UserData"] = base64.StdEncoding.EncodeToString([]byte(data))
		}
	}

	// vpc
	if v, ok := d.GetOk("vpc_id"); ok {
		vpcId := v.(string)
		params["VirtualPrivateCloud.VpcId"] = vpcId
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		subnetId := v.(string)
		params["VirtualPrivateCloud.SubnetId"] = subnetId
	}
	if v, ok := d.GetOk("private_ip"); ok {
		ip := v.(string)
		params["VirtualPrivateCloud.PrivateIpAddresses.0"] = ip
	}

	// tag
	if v, ok := d.GetOk("tags"); ok {
		i := 0
		params["TagSpecification.0.ResourceType"] = "instance"
		for key, value := range v.(map[string]interface{}) {
			params["TagSpecification.0.Tags."+strconv.Itoa(i)+".Key"] = key
			params["TagSpecification.0.Tags."+strconv.Itoa(i)+".Value"] = value.(string)
		}
		i = i + 1
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		response, err := client.SendRequest("cvm", params)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		var jsonresp struct {
			Response struct {
				Error struct {
					Code    string `json:"Code"`
					Message string `json:"Message"`
				}
				InstanceIdSet []string
				RequestId     string
			}
		}

		err = json.Unmarshal([]byte(response), &jsonresp)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if jsonresp.Response.Error.Code == "VpcIpIsUsed" {
			return resource.RetryableError(fmt.Errorf("error: %v, request id: %v", jsonresp.Response.Error.Message, jsonresp.Response.RequestId))
		}

		if jsonresp.Response.Error.Code != "" {
			err = fmt.Errorf(
				"tencentcloud_instance got error, code:%v, message:%v, request id:%v",
				jsonresp.Response.Error.Code,
				jsonresp.Response.Error.Message,
				jsonresp.Response.RequestId,
			)
			return resource.NonRetryableError(err)
		}

		if len(jsonresp.Response.InstanceIdSet) == 0 {
			err = fmt.Errorf("tencentcloud_instance no instance id returned")
			return resource.NonRetryableError(err)
		}

		var instanceStatusMap map[string]string
		instanceStatusMap, err = waitInstanceReachTargetStatus(client, jsonresp.Response.InstanceIdSet, "RUNNING")
		if err != nil {
			return resource.NonRetryableError(err)
		}

		instanceId := jsonresp.Response.InstanceIdSet[0]
		d.SetId(instanceId)
		d.Set("instance_status", instanceStatusMap[instanceId])
		d.Set("data_disks", dataDisksAttr)

		return nil
	})

	if err != nil {
		return err
	}

	return resourceTencentCloudInstanceRead(d, m)
}

func resourceTencentCloudInstanceRead(d *schema.ResourceData, m interface{}) error {
	instanceId := d.Id()
	params := map[string]string{
		"Version": "2017-03-12",
		"Action":  "DescribeInstances",
	}
	params["InstanceIds.0"] = instanceId

	client := m.(*TencentCloudClient).commonConn
	response, err := client.SendRequest("cvm", params)
	if err != nil {
		return err
	}

	var jsonresp struct {
		Response struct {
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			}
			TotalCount  int
			InstanceSet []struct {
				Placement struct {
					Zone      string   `json:"Zone"`
					ProjectId int      `json:"ProjectId"`
					HostIds   []string `json:"HostIds"`
				} `json:"Placement"`

				InstanceId   string `json:"InstanceId"`
				InstanceType string `json:"InstanceType"`
				CPU          int    `json:"CPU"`
				Memory       int    `json:"Memory"`

				RestrictState      string `json:"RestrictState"`
				InstanceName       string `json:"InstanceName"`
				InstanceChargeType string `json:"InstanceChargeType"`
				InstanceState      string `json:"InstanceState"`

				SystemDisk struct {
					DiskType string `json:"DiskType"`
					DiskId   string `json:"DiskId"`
					DiskSize int    `json:"DiskSize"`
				} `json:"SystemDisk"`

				DataDisks []struct {
					DiskType           string `json:"DiskType"`
					DiskId             string `json:"DiskId"`
					DiskSize           int    `json:"DiskSize"`
					DeleteWithInstance bool   `json:"DeleteWithInstance"`
				} `json:"DataDisks"`

				PrivateIpAddresses []string `json:"PrivateIpAddresses"`
				PublicIpAddresses  []string `json:"PublicIpAddresses"`

				InternetAccessible struct {
					InternetChargeType      string `json:"InternetChargeType"`
					InternetMaxBandwidthOut int    `json:"InternetMaxBandwidthOut"`
					PublicIpAssigned        bool   `json:"PublicIpAssigned"`
				} `json:"InternetAccessible"`

				VirtualPrivateCloud struct {
					VpcId              string   `json:"VpcId"`
					SubnetId           string   `json:"SubnetId"`
					AsVpcGateway       bool     `json:"AsVpcGateway"`
					PrivateIpAddresses []string `json:"PrivateIpAddresses"`
				}

				SecurityGroupIds []string `json:"SecurityGroupIds"`

				LoginSettings struct {
					KeyIds []string `json:"KeyIds"`
				} `json:"LoginSettings"`

				ImageId string `json:"ImageId"`

				RenewFlag   string    `json:"RenewFlag"`
				CreatedTime time.Time `json:"CreatedTime"`
				ExpiredTime time.Time `json:"ExpiredTime"`

				Tags []struct {
					Key   string `json:"Key"`
					Value string `json:"Value"`
				} `json:"Tags"`
			} `json:"InstanceSet"`
			RequestId string
		}
	}
	err = json.Unmarshal([]byte(response), &jsonresp)
	if err != nil {
		return err
	}
	if jsonresp.Response.Error.Code != "" {
		return fmt.Errorf(
			"tencentcloud_instance got error, code:%v, message:%v, request id:%v",
			jsonresp.Response.Error.Code,
			jsonresp.Response.Error.Message,
			jsonresp.Response.RequestId,
		)
	}
	if len(jsonresp.Response.InstanceSet) == 0 {
		d.SetId("")
		return nil
	}
	privateIPs := jsonresp.Response.InstanceSet[0].PrivateIpAddresses
	if len(privateIPs) > 0 {
		d.Set("private_ip", privateIPs[0])
	}
	publicIPs := jsonresp.Response.InstanceSet[0].PublicIpAddresses
	if len(publicIPs) > 0 {
		d.Set("public_ip", publicIPs[0])
	} else {
		d.Set("public_ip", "")
	}
	systemDiskType := jsonresp.Response.InstanceSet[0].SystemDisk.DiskType
	d.Set("system_disk_type", systemDiskType)
	systemDiskSize := jsonresp.Response.InstanceSet[0].SystemDisk.DiskSize
	d.Set("system_disk_size", systemDiskSize)

	ImageId := jsonresp.Response.InstanceSet[0].ImageId
	d.Set("image_id", ImageId)
	InstanceName := jsonresp.Response.InstanceSet[0].InstanceName
	d.Set("instance_name", InstanceName)
	InstanceType := jsonresp.Response.InstanceSet[0].InstanceType
	d.Set("instance_type", InstanceType)
	InstanceState := jsonresp.Response.InstanceSet[0].InstanceState
	d.Set("instance_status", InstanceState)
	Zone := jsonresp.Response.InstanceSet[0].Placement.Zone
	d.Set("availability_zone", Zone)
	InternetMaxBandwidthOut := jsonresp.Response.InstanceSet[0].InternetAccessible.InternetMaxBandwidthOut
	d.Set("internet_max_bandwidth_out", InternetMaxBandwidthOut)
	d.Set("allocate_public_ip", len(publicIPs) > 0)

	var dataDiskList []map[string]interface{}
	dataDisks := jsonresp.Response.InstanceSet[0].DataDisks
	for _, dataDisk := range dataDisks {
		m := make(map[string]interface{})
		diskType := dataDisk.DiskType
		diskSize := dataDisk.DiskSize
		deleteWithInstance := dataDisk.DeleteWithInstance
		m["data_disk_type"] = diskType
		m["data_disk_size"] = diskSize
		m["delete_with_instance"] = deleteWithInstance
		dataDiskList = append(dataDiskList, m)
	}
	d.Set("data_disks", dataDiskList)

	securityGroupIds := jsonresp.Response.InstanceSet[0].SecurityGroupIds
	if len(securityGroupIds) > 0 {
		d.Set("security_groups", securityGroupIds)
	}

	loginSettings := jsonresp.Response.InstanceSet[0].LoginSettings
	keyIds := loginSettings.KeyIds
	if len(keyIds) > 0 {
		keyName := keyIds[0]
		d.Set("key_name", keyName)
	}
	if v, ok := d.GetOk("password"); ok {
		passwd := v.(string)
		d.Set("password", passwd)
	}

	virtualPrivateCloud := jsonresp.Response.InstanceSet[0].VirtualPrivateCloud
	vpcId := virtualPrivateCloud.VpcId
	if len(vpcId) > 0 {
		d.Set("vpc_id", vpcId)
	}
	subnetId := virtualPrivateCloud.SubnetId
	if len(subnetId) > 0 {
		d.Set("subnet_id", subnetId)
	}

	// we do not allow to modify it, so we do not make it computed as well
	//tags := make(map[string]string)
	//for _, tag := range jsonresp.Response.InstanceSet[0].Tags {
	//	tags[tag.Key] = tag.Value
	//}
	//d.Set("tags", tags)

	return nil
}

func resourceTencentCloudInstanceUpdate(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*TencentCloudClient).commonConn
	instanceId := d.Id()

	for _, field := range unsupportedUpdateFields {
		if d.HasChange(field) {
			return fmt.Errorf("tencentcloud_instance update on %v is not supported yet", field)
		}
	}

	d.Partial(true)

	var isUpdatingImage bool

	if d.HasChange("image_id") {
		isUpdatingImage = true
	}

	if d.HasChange("instance_name") {
		d.SetPartial("instance_name")
		oldInstanceName, newInstanceName := d.GetChange("instance_name")
		log.Printf("[DEBUG] tencentcloud_instance rename instance_name from %v to %v", oldInstanceName, newInstanceName)
		err = renameInstancesName(client, []string{instanceId}, newInstanceName.(string))
		if err != nil {
			return err
		}
	}

	if d.HasChange("project_id") {
		oldProjectId, newProjectId := d.GetChange("project_id")
		log.Printf("[DEBUG] tencentcloud_instance modify project_id from %v to %v", oldProjectId, newProjectId)
		err = modifyInstancesProject(client, []string{instanceId}, newProjectId.(int))
		if err != nil {
			return err
		}
		d.SetPartial("project_id")
	}

	if d.HasChange("key_name") {
		d.SetPartial("key_name")
		if isUpdatingImage {
			goto LABEL_REINSTALL
		}
		oldKey, newKey := d.GetChange("key_name")
		log.Printf("[DEBUG] tencentcloud_instance rebind key pair, old key: %v, new key: %v", oldKey, newKey)

		_, err := waitInstanceReachOneOfTargetStatusList(
			client,
			[]string{instanceId},
			[]string{
				"STOPPED",
				"RUNNING",
			},
		)
		if err != nil {
			return err
		}

		err = bindKeyPiar(client, instanceId, newKey.(string))
		if err != nil {
			return err
		}
	}

	if d.HasChange("password") {
		d.SetPartial("password")
		if isUpdatingImage {
			goto LABEL_REINSTALL
		}
		_, newValue := d.GetChange("password")
		log.Printf("[DEBUG] tencentcloud_instance reset password\n")
		err = resetInstancePassword(client, instanceId, newValue.(string))
		if err != nil {
			return err
		}
	}

	if d.HasChange("security_groups") {
		_, n := d.GetChange("security_groups")
		ns := n.(*schema.Set)

		sgIds := expandStringList(ns.List())
		if len(sgIds) == 0 {
			return fmt.Errorf("tencentcloud_instance security groups are not allow to be empty")
		}

		err = bindInstanceWithSgIds(client, d.Id(), sgIds)
		if err != nil {
			return err
		}
		d.SetPartial("security_groups")
	}

LABEL_REINSTALL:
	if d.HasChange("image_id") {
		d.SetPartial("image_id")
		oldValue, newValue := d.GetChange("image_id")
		log.Printf("[DEBUG] tencentcloud_instance reinstall image from %v to %v\n", oldValue, newValue)

		err = resetInstanceSystem(client, d, instanceId, newValue.(string))
		if err != nil {
			return err
		}
	}

	d.Partial(false)

	return resourceTencentCloudInstanceRead(d, m)
}

func resourceTencentCloudInstanceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*TencentCloudClient).commonConn

	params := map[string]string{
		"Version":       "2017-03-12",
		"Action":        "TerminateInstances",
		"InstanceIds.0": d.Id(),
	}

	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		response, err := client.SendRequest("cvm", params)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		var jsonresp struct {
			Response struct {
				Error struct {
					Code    string `json:"Code"`
					Message string `json:"Message"`
				}
				RequestId string
			}
		}
		err = json.Unmarshal([]byte(response), &jsonresp)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if jsonresp.Response.Error.Code == "InternalError" {
			return resource.RetryableError(fmt.Errorf("error: %v, request id: %v", jsonresp.Response.Error.Message, jsonresp.Response.RequestId))
		}
		return nil
	})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
