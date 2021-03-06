package ibm

import (
	"fmt"
	"log"
	"time"

	"github.com/IBM/vpc-go-sdk/vpcclassicv1"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	isVPNGatewayConnectionAdminStateup              = "admin_state_up"
	isVPNGatewayConnectionName                      = "name"
	isVPNGatewayConnectionVPNGateway                = "vpn_gateway"
	isVPNGatewayConnectionPeerAddress               = "peer_address"
	isVPNGatewayConnectionPreSharedKey              = "preshared_key"
	isVPNGatewayConnectionLocalCIDRS                = "local_cidrs"
	isVPNGatewayConnectionPeerCIDRS                 = "peer_cidrs"
	isVPNGatewayConnectionIKEPolicy                 = "ike_policy"
	isVPNGatewayConnectionIPSECPolicy               = "ipsec_policy"
	isVPNGatewayConnectionDeadPeerDetectionAction   = "action"
	isVPNGatewayConnectionDeadPeerDetectionInterval = "interval"
	isVPNGatewayConnectionDeadPeerDetectionTimeout  = "timeout"
	isVPNGatewayConnectionStatus                    = "status"
	isVPNGatewayConnectionDeleting                  = "deleting"
	isVPNGatewayConnectionDeleted                   = "done"
	isVPNGatewayConnectionProvisioning              = "provisioning"
	isVPNGatewayConnectionProvisioningDone          = "done"
)

func resourceIBMISVPNGatewayConnection() *schema.Resource {
	return &schema.Resource{
		Create:   resourceIBMISVPNGatewayConnectionCreate,
		Read:     resourceIBMISVPNGatewayConnectionRead,
		Update:   resourceIBMISVPNGatewayConnectionUpdate,
		Delete:   resourceIBMISVPNGatewayConnectionDelete,
		Exists:   resourceIBMISVPNGatewayConnectionExists,
		Importer: &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

			isVPNGatewayConnectionName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateISName,
				Description:  "VPN Gateway connection name",
			},

			isVPNGatewayConnectionVPNGateway: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "VPN Gateway info",
			},

			isVPNGatewayConnectionPeerAddress: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "VPN gateway connection peer address",
			},

			isVPNGatewayConnectionPreSharedKey: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "vpn gateway",
			},

			isVPNGatewayConnectionAdminStateup: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "VPN gateway connection admin state",
			},

			isVPNGatewayConnectionLocalCIDRS: {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "VPN gateway connection local CIDRs",
			},

			isVPNGatewayConnectionPeerCIDRS: {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "VPN gateway connection peer CIDRs",
			},

			isVPNGatewayConnectionDeadPeerDetectionAction: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validateAllowedStringValue([]string{"restart", "clear", "hold", "none"}),
				Description:  "Action detection for dead peer detection action",
			},
			isVPNGatewayConnectionDeadPeerDetectionInterval: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validateDeadPeerDetectionInterval,
				Description:  "Interval for dead peer detection interval",
			},
			isVPNGatewayConnectionDeadPeerDetectionTimeout: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      120,
				ValidateFunc: validateDeadPeerDetectionInterval,
				Description:  "Timeout for dead peer detection",
			},

			isVPNGatewayConnectionIPSECPolicy: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP security policy for vpn gateway connection",
			},

			isVPNGatewayConnectionIKEPolicy: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VPN gateway connection IKE Policy",
			},

			isVPNGatewayConnectionStatus: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "VPN gateway connection status",
			},
		},
	}
}

func resourceIBMISVPNGatewayConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	userDetails, err := meta.(ClientSession).BluemixUserDetails()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] VPNGatewayConnection create")
	name := d.Get(isVPNGatewayConnectionName).(string)
	gatewayID := d.Get(isVPNGatewayConnectionVPNGateway).(string)
	peerAddress := d.Get(isVPNGatewayConnectionPeerAddress).(string)
	prephasedKey := d.Get(isVPNGatewayConnectionPreSharedKey).(string)

	stateUp := false
	if _, ok := d.GetOk(isVPNGatewayConnectionAdminStateup); ok {
		stateUp = d.Get(isVPNGatewayConnectionAdminStateup).(bool)
	}
	var interval, timeout int64
	if intvl, ok := d.GetOk(isVPNGatewayConnectionDeadPeerDetectionInterval); ok {
		interval = int64(intvl.(int))
	} else {
		interval = 30
	}

	if tout, ok := d.GetOk(isVPNGatewayConnectionDeadPeerDetectionTimeout); ok {
		timeout = int64(tout.(int))
	} else {
		timeout = 120
	}
	var action string
	if act, ok := d.GetOk(isVPNGatewayConnectionDeadPeerDetectionAction); ok {
		action = act.(string)
	} else {
		action = "none"
	}

	if userDetails.generation == 1 {
		err := classicVpngwconCreate(d, meta, name, gatewayID, peerAddress, prephasedKey, action, interval, timeout, stateUp)
		if err != nil {
			return err
		}
	} else {
		err := vpngwconCreate(d, meta, name, gatewayID, peerAddress, prephasedKey, action, interval, timeout, stateUp)
		if err != nil {
			return err
		}
	}
	return resourceIBMISVPNGatewayConnectionRead(d, meta)
}

func classicVpngwconCreate(d *schema.ResourceData, meta interface{}, name, gatewayID, peerAddress, prephasedKey, action string, interval, timeout int64, stateUp bool) error {
	sess, err := classicVpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcclassicv1.CreateVPNGatewayConnectionOptions{
		VPNGatewayID: &gatewayID,
		PeerAddress:  &peerAddress,
		Psk:          &prephasedKey,
		AdminStateUp: &stateUp,
		DeadPeerDetection: &vpcclassicv1.VPNGatewayConnectionDpdPrototype{
			Action:   &action,
			Interval: &interval,
			Timeout:  &timeout,
		},
		Name: &name,
	}

	if _, ok := d.GetOk(isVPNGatewayConnectionLocalCIDRS); ok {
		localCidrs := expandStringList((d.Get(isVPNGatewayConnectionLocalCIDRS).(*schema.Set)).List())
		options.LocalCidrs = localCidrs
	}
	if _, ok := d.GetOk(isVPNGatewayConnectionPeerCIDRS); ok {
		peerCidrs := expandStringList((d.Get(isVPNGatewayConnectionPeerCIDRS).(*schema.Set)).List())
		options.PeerCidrs = peerCidrs
	}

	var ikePolicyIdentity, ipsecPolicyIdentity string

	if ikePolicy, ok := d.GetOk(isVPNGatewayConnectionIKEPolicy); ok {
		ikePolicyIdentity = ikePolicy.(string)
		options.IkePolicy = &vpcclassicv1.IkePolicyIdentity{
			ID: &ikePolicyIdentity,
		}
	}
	if ipsecPolicy, ok := d.GetOk(isVPNGatewayConnectionIPSECPolicy); ok {
		ipsecPolicyIdentity = ipsecPolicy.(string)
		options.IpsecPolicy = &vpcclassicv1.IPsecPolicyIdentity{
			ID: &ipsecPolicyIdentity,
		}
	}

	vpnGatewayConnection, response, err := sess.CreateVPNGatewayConnection(options)
	if err != nil {
		return fmt.Errorf("[DEBUG] Create VPN Gateway Connection err %s\n%s", err, response)
	}
	d.SetId(fmt.Sprintf("%s/%s", gatewayID, *vpnGatewayConnection.ID))
	log.Printf("[INFO] VPNGatewayConnection : %s/%s", gatewayID, *vpnGatewayConnection.ID)
	return nil
}

func vpngwconCreate(d *schema.ResourceData, meta interface{}, name, gatewayID, peerAddress, prephasedKey, action string, interval, timeout int64, stateUp bool) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcv1.CreateVPNGatewayConnectionOptions{
		VPNGatewayID: &gatewayID,
		PeerAddress:  &peerAddress,
		Psk:          &prephasedKey,
		AdminStateUp: &stateUp,
		DeadPeerDetection: &vpcv1.VPNGatewayConnectionDpdPrototype{
			Action:   &action,
			Interval: &interval,
			Timeout:  &timeout,
		},
		Name: &name,
	}

	if _, ok := d.GetOk(isVPNGatewayConnectionLocalCIDRS); ok {
		localCidrs := expandStringList((d.Get(isVPNGatewayConnectionLocalCIDRS).(*schema.Set)).List())
		options.LocalCidrs = localCidrs
	}
	if _, ok := d.GetOk(isVPNGatewayConnectionPeerCIDRS); ok {
		peerCidrs := expandStringList((d.Get(isVPNGatewayConnectionPeerCIDRS).(*schema.Set)).List())
		options.PeerCidrs = peerCidrs
	}

	var ikePolicyIdentity, ipsecPolicyIdentity string

	if ikePolicy, ok := d.GetOk(isVPNGatewayConnectionIKEPolicy); ok {
		ikePolicyIdentity = ikePolicy.(string)
		options.IkePolicy = &vpcv1.IkePolicyIdentity{
			ID: &ikePolicyIdentity,
		}
	} else {
		options.IkePolicy = nil
	}
	if ipsecPolicy, ok := d.GetOk(isVPNGatewayConnectionIPSECPolicy); ok {
		ipsecPolicyIdentity = ipsecPolicy.(string)
		options.IpsecPolicy = &vpcv1.IPsecPolicyIdentity{
			ID: &ipsecPolicyIdentity,
		}
	} else {
		options.IpsecPolicy = nil
	}

	vpnGatewayConnection, response, err := sess.CreateVPNGatewayConnection(options)
	if err != nil {
		return fmt.Errorf("[DEBUG] Create VPN Gateway Connection err %s\n%s", err, response)
	}
	d.SetId(fmt.Sprintf("%s/%s", gatewayID, *vpnGatewayConnection.ID))
	log.Printf("[INFO] VPNGatewayConnection : %s/%s", gatewayID, *vpnGatewayConnection.ID)
	return nil
}

func resourceIBMISVPNGatewayConnectionRead(d *schema.ResourceData, meta interface{}) error {
	userDetails, err := meta.(ClientSession).BluemixUserDetails()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	gID := parts[0]
	gConnID := parts[1]

	if userDetails.generation == 1 {
		err := classicVpngwconGet(d, meta, gID, gConnID)
		if err != nil {
			return err
		}
	} else {
		err := vpngwconGet(d, meta, gID, gConnID)
		if err != nil {
			return err
		}
	}
	return nil
}

func classicVpngwconGet(d *schema.ResourceData, meta interface{}, gID, gConnID string) error {
	sess, err := classicVpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcclassicv1.GetVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	vpnGatewayConnection, response, err := sess.GetVPNGatewayConnection(options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Getting Vpn Gateway Connection (%s): %s\n%s", gConnID, err, response)
	}
	d.Set(isVPNGatewayConnectionName, *vpnGatewayConnection.Name)
	d.Set(isVPNGatewayConnectionVPNGateway, gID)
	d.Set(isVPNGatewayConnectionAdminStateup, *vpnGatewayConnection.AdminStateUp)
	d.Set(isVPNGatewayConnectionPeerAddress, *vpnGatewayConnection.PeerAddress)
	d.Set(isVPNGatewayConnectionPreSharedKey, *vpnGatewayConnection.Psk)
	d.Set(isVPNGatewayConnectionLocalCIDRS, flattenStringList(vpnGatewayConnection.LocalCidrs))
	d.Set(isVPNGatewayConnectionPeerCIDRS, flattenStringList(vpnGatewayConnection.PeerCidrs))
	if vpnGatewayConnection.IkePolicy != nil {
		ikepolicy := vpnGatewayConnection.IkePolicy.(*vpcclassicv1.IkePolicyIdentity)
		d.Set(isVPNGatewayConnectionIKEPolicy, *ikepolicy.ID)
	}
	if vpnGatewayConnection.IpsecPolicy != nil {
		ipsecpolicy := vpnGatewayConnection.IpsecPolicy.(*vpcclassicv1.IPsecPolicyIdentity)
		d.Set(isVPNGatewayConnectionIPSECPolicy, *ipsecpolicy.ID)
	}
	d.Set(isVPNGatewayConnectionDeadPeerDetectionAction, *vpnGatewayConnection.DeadPeerDetection.Action)
	d.Set(isVPNGatewayConnectionDeadPeerDetectionInterval, *vpnGatewayConnection.DeadPeerDetection.Interval)
	d.Set(isVPNGatewayConnectionDeadPeerDetectionTimeout, *vpnGatewayConnection.DeadPeerDetection.Timeout)
	return nil
}

func vpngwconGet(d *schema.ResourceData, meta interface{}, gID, gConnID string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	options := &vpcv1.GetVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	vpnGatewayConnection, response, err := sess.GetVPNGatewayConnection(options)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Getting Vpn Gateway Connection (%s): %s\n%s", gConnID, err, response)
	}
	d.Set(isVPNGatewayConnectionName, *vpnGatewayConnection.Name)
	d.Set(isVPNGatewayConnectionVPNGateway, gID)
	d.Set(isVPNGatewayConnectionAdminStateup, *vpnGatewayConnection.AdminStateUp)
	d.Set(isVPNGatewayConnectionPeerAddress, *vpnGatewayConnection.PeerAddress)
	d.Set(isVPNGatewayConnectionPreSharedKey, *vpnGatewayConnection.Psk)
	d.Set(isVPNGatewayConnectionLocalCIDRS, flattenStringList(vpnGatewayConnection.LocalCidrs))
	d.Set(isVPNGatewayConnectionPeerCIDRS, flattenStringList(vpnGatewayConnection.PeerCidrs))
	if vpnGatewayConnection.IkePolicy != nil {
		ikepolicy := vpnGatewayConnection.IkePolicy.(*vpcv1.IkePolicyIdentity)
		d.Set(isVPNGatewayConnectionIKEPolicy, *ikepolicy.ID)
	}
	if vpnGatewayConnection.IpsecPolicy != nil {
		ipsecpolicy := vpnGatewayConnection.IpsecPolicy.(*vpcv1.IPsecPolicyIdentity)
		d.Set(isVPNGatewayConnectionIPSECPolicy, *ipsecpolicy.ID)
	}
	d.Set(isVPNGatewayConnectionDeadPeerDetectionAction, *vpnGatewayConnection.DeadPeerDetection.Action)
	d.Set(isVPNGatewayConnectionDeadPeerDetectionInterval, *vpnGatewayConnection.DeadPeerDetection.Interval)
	d.Set(isVPNGatewayConnectionDeadPeerDetectionTimeout, *vpnGatewayConnection.DeadPeerDetection.Timeout)
	return nil
}

func resourceIBMISVPNGatewayConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	userDetails, err := meta.(ClientSession).BluemixUserDetails()
	if err != nil {
		return err
	}
	hasChanged := false

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	gID := parts[0]
	gConnID := parts[1]

	if userDetails.generation == 1 {
		err := classicVpngwconUpdate(d, meta, gID, gConnID, hasChanged)
		if err != nil {
			return err
		}
	} else {
		err := vpngwconUpdate(d, meta, gID, gConnID, hasChanged)
		if err != nil {
			return err
		}
	}
	return resourceIBMISVPNGatewayConnectionRead(d, meta)
}

func classicVpngwconUpdate(d *schema.ResourceData, meta interface{}, gID, gConnID string, hasChanged bool) error {
	sess, err := classicVpcClient(meta)
	if err != nil {
		return err
	}

	updateVpnGatewayConnectionOptions := &vpcclassicv1.UpdateVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	if d.HasChange(isVPNGatewayConnectionName) {
		name := d.Get(isVPNGatewayConnectionName).(string)
		updateVpnGatewayConnectionOptions.Name = &name
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionPeerAddress) {
		peerAddress := d.Get(isVPNGatewayConnectionPeerAddress).(string)
		updateVpnGatewayConnectionOptions.PeerAddress = &peerAddress
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionPreSharedKey) {
		psk := d.Get(isVPNGatewayConnectionPreSharedKey).(string)
		updateVpnGatewayConnectionOptions.Psk = &psk
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionDeadPeerDetectionAction) || d.HasChange(isVPNGatewayConnectionDeadPeerDetectionInterval) || d.HasChange(isVPNGatewayConnectionDeadPeerDetectionTimeout) {
		action := d.Get(isVPNGatewayConnectionDeadPeerDetectionAction).(string)
		interval := int64(d.Get(isVPNGatewayConnectionDeadPeerDetectionInterval).(int))
		timeout := int64(d.Get(isVPNGatewayConnectionDeadPeerDetectionTimeout).(int))
		updateVpnGatewayConnectionOptions.DeadPeerDetection.Action = &action
		updateVpnGatewayConnectionOptions.DeadPeerDetection.Interval = &interval
		updateVpnGatewayConnectionOptions.DeadPeerDetection.Timeout = &timeout
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionIKEPolicy) {
		ikePolicyIdentity := d.Get(isVPNGatewayConnectionIKEPolicy).(string)
		updateVpnGatewayConnectionOptions.IkePolicy = &vpcclassicv1.IkePolicyIdentity{
			ID: &ikePolicyIdentity,
		}
		hasChanged = true
	} else {
		updateVpnGatewayConnectionOptions.IkePolicy = nil
	}

	if d.HasChange(isVPNGatewayConnectionIPSECPolicy) {
		ipsecPolicyIdentity := d.Get(isVPNGatewayConnectionIPSECPolicy).(string)
		updateVpnGatewayConnectionOptions.IpsecPolicy = &vpcclassicv1.IPsecPolicyIdentity{
			ID: &ipsecPolicyIdentity,
		}
		hasChanged = true
	} else {
		updateVpnGatewayConnectionOptions.IpsecPolicy = nil
	}

	if d.HasChange(isVPNGatewayConnectionAdminStateup) {
		adminStateUp := d.Get(isVPNGatewayConnectionAdminStateup).(bool)
		updateVpnGatewayConnectionOptions.AdminStateUp = &adminStateUp
		hasChanged = true
	}

	if hasChanged {
		_, response, err := sess.UpdateVPNGatewayConnection(updateVpnGatewayConnectionOptions)
		if err != nil {
			return fmt.Errorf("Error updating Vpn Gateway Connection: %s\n%s", err, response)
		}
	}
	return nil
}

func vpngwconUpdate(d *schema.ResourceData, meta interface{}, gID, gConnID string, hasChanged bool) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}

	updateVpnGatewayConnectionOptions := &vpcv1.UpdateVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	if d.HasChange(isVPNGatewayConnectionName) {
		name := d.Get(isVPNGatewayConnectionName).(string)
		updateVpnGatewayConnectionOptions.Name = &name
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionPeerAddress) {
		peerAddress := d.Get(isVPNGatewayConnectionPeerAddress).(string)
		updateVpnGatewayConnectionOptions.PeerAddress = &peerAddress
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionPreSharedKey) {
		psk := d.Get(isVPNGatewayConnectionPreSharedKey).(string)
		updateVpnGatewayConnectionOptions.Psk = &psk
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionDeadPeerDetectionAction) || d.HasChange(isVPNGatewayConnectionDeadPeerDetectionInterval) || d.HasChange(isVPNGatewayConnectionDeadPeerDetectionTimeout) {
		action := d.Get(isVPNGatewayConnectionDeadPeerDetectionAction).(string)
		interval := int64(d.Get(isVPNGatewayConnectionDeadPeerDetectionInterval).(int))
		timeout := int64(d.Get(isVPNGatewayConnectionDeadPeerDetectionTimeout).(int))
		updateVpnGatewayConnectionOptions.DeadPeerDetection.Action = &action
		updateVpnGatewayConnectionOptions.DeadPeerDetection.Interval = &interval
		updateVpnGatewayConnectionOptions.DeadPeerDetection.Timeout = &timeout
		hasChanged = true
	}

	if d.HasChange(isVPNGatewayConnectionIKEPolicy) {
		ikePolicyIdentity := d.Get(isVPNGatewayConnectionIKEPolicy).(string)
		updateVpnGatewayConnectionOptions.IkePolicy = &vpcv1.IkePolicyIdentity{
			ID: &ikePolicyIdentity,
		}
		hasChanged = true
	} else {
		updateVpnGatewayConnectionOptions.IkePolicy = nil
	}

	if d.HasChange(isVPNGatewayConnectionIPSECPolicy) {
		ipsecPolicyIdentity := d.Get(isVPNGatewayConnectionIPSECPolicy).(string)
		updateVpnGatewayConnectionOptions.IpsecPolicy = &vpcv1.IPsecPolicyIdentity{
			ID: &ipsecPolicyIdentity,
		}
		hasChanged = true
	} else {
		updateVpnGatewayConnectionOptions.IpsecPolicy = nil
	}

	if d.HasChange(isVPNGatewayConnectionAdminStateup) {
		adminStateUp := d.Get(isVPNGatewayConnectionAdminStateup).(bool)
		updateVpnGatewayConnectionOptions.AdminStateUp = &adminStateUp
		hasChanged = true
	}

	if hasChanged {
		_, response, err := sess.UpdateVPNGatewayConnection(updateVpnGatewayConnectionOptions)
		if err != nil {
			return fmt.Errorf("Error updating Vpn Gateway Connection: %s\n%s", err, response)
		}
	}
	return nil
}

func resourceIBMISVPNGatewayConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	userDetails, err := meta.(ClientSession).BluemixUserDetails()
	if err != nil {
		return err
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return err
	}

	gID := parts[0]
	gConnID := parts[1]

	if userDetails.generation == 1 {
		err := classicVpngwconDelete(d, meta, gID, gConnID)
		if err != nil {
			return err
		}
	} else {
		err := vpngwconDelete(d, meta, gID, gConnID)
		if err != nil {
			return err
		}
	}
	return nil
}

func classicVpngwconDelete(d *schema.ResourceData, meta interface{}, gID, gConnID string) error {
	sess, err := classicVpcClient(meta)
	if err != nil {
		return err
	}
	getVpnGatewayConnectionOptions := &vpcclassicv1.GetVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	_, response, err := sess.GetVPNGatewayConnection(getVpnGatewayConnectionOptions)

	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Getting Vpn Gateway Connection(%s): %s\n%s", gConnID, err, response)
	}
	deleteVpnGatewayConnectionOptions := &vpcclassicv1.DeleteVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	response, err = sess.DeleteVPNGatewayConnection(deleteVpnGatewayConnectionOptions)
	if err != nil {
		return fmt.Errorf("Error Deleting Vpn Gateway Connection : %s\n%s", err, response)
	}

	_, err = isWaitForClassicVPNGatewayConnectionDeleted(sess, gID, gConnID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf(
			"Error checking for Vpn Gateway Connection (%s) is deleted: %s", gConnID, err)
	}

	d.SetId("")
	return nil
}

func vpngwconDelete(d *schema.ResourceData, meta interface{}, gID, gConnID string) error {
	sess, err := vpcClient(meta)
	if err != nil {
		return err
	}
	getVpnGatewayConnectionOptions := &vpcv1.GetVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	_, response, err := sess.GetVPNGatewayConnection(getVpnGatewayConnectionOptions)

	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Getting Vpn Gateway Connection(%s): %s\n%s", gConnID, err, response)
	}

	deleteVpnGatewayConnectionOptions := &vpcv1.DeleteVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	response, err = sess.DeleteVPNGatewayConnection(deleteVpnGatewayConnectionOptions)
	if err != nil {
		return fmt.Errorf("Error Deleting Vpn Gateway Connection : %s\n%s", err, response)
	}

	_, err = isWaitForVPNGatewayConnectionDeleted(sess, gID, gConnID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf(
			"Error checking for Vpn Gateway Connection (%s) is deleted: %s", gConnID, err)
	}

	d.SetId("")
	return nil
}

func isWaitForClassicVPNGatewayConnectionDeleted(vpnGatewayConnection *vpcclassicv1.VpcClassicV1, gID, gConnID string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for VPNGatewayConnection (%s) to be deleted.", gConnID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", isVPNGatewayConnectionDeleting},
		Target:     []string{"", isVPNGatewayConnectionDeleted},
		Refresh:    isClassicVPNGatewayConnectionDeleteRefreshFunc(vpnGatewayConnection, gID, gConnID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func isClassicVPNGatewayConnectionDeleteRefreshFunc(vpnGatewayConnection *vpcclassicv1.VpcClassicV1, gID, gConnID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getVpnGatewayConnectionOptions := &vpcclassicv1.GetVPNGatewayConnectionOptions{
			VPNGatewayID: &gID,
			ID:           &gConnID,
		}
		vpngwcon, response, err := vpnGatewayConnection.GetVPNGatewayConnection(getVpnGatewayConnectionOptions)
		if err != nil {
			if response != nil && response.StatusCode == 404 {
				return vpngwcon, isVPNGatewayConnectionDeleted, nil
			}
			return nil, "", fmt.Errorf("The Vpn Gateway Connection %s failed to delete: %s\n%s", gConnID, err, response)
		}
		return vpngwcon, isVPNGatewayConnectionDeleting, nil
	}
}

func isWaitForVPNGatewayConnectionDeleted(vpnGatewayConnection *vpcv1.VpcV1, gID, gConnID string, timeout time.Duration) (interface{}, error) {
	log.Printf("Waiting for VPNGatewayConnection (%s) to be deleted.", gConnID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"retry", isVPNGatewayConnectionDeleting},
		Target:     []string{"", isVPNGatewayConnectionDeleted},
		Refresh:    isVPNGatewayConnectionDeleteRefreshFunc(vpnGatewayConnection, gID, gConnID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func isVPNGatewayConnectionDeleteRefreshFunc(vpnGatewayConnection *vpcv1.VpcV1, gID, gConnID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getVpnGatewayConnectionOptions := &vpcv1.GetVPNGatewayConnectionOptions{
			VPNGatewayID: &gID,
			ID:           &gConnID,
		}
		vpngwcon, response, err := vpnGatewayConnection.GetVPNGatewayConnection(getVpnGatewayConnectionOptions)
		if err != nil {
			if response != nil && response.StatusCode == 404 {
				return vpngwcon, isVPNGatewayConnectionDeleted, nil
			}
			return nil, "", fmt.Errorf("The Vpn Gateway Connection %s failed to delete: %s\n%s", gConnID, err, response)
		}
		return vpngwcon, isVPNGatewayConnectionDeleting, nil
	}
}

func resourceIBMISVPNGatewayConnectionExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	userDetails, err := meta.(ClientSession).BluemixUserDetails()
	if err != nil {
		return false, err
	}
	parts, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	gID := parts[0]
	gConnID := parts[1]
	if userDetails.generation == 1 {
		exists, err := classicVpngwconExists(d, meta, gID, gConnID)
		return exists, err
	} else {
		exists, err := vpngwconExists(d, meta, gID, gConnID)
		return exists, err
	}
}

func classicVpngwconExists(d *schema.ResourceData, meta interface{}, gID, gConnID string) (bool, error) {
	sess, err := classicVpcClient(meta)
	if err != nil {
		return false, err
	}

	getVpnGatewayConnectionOptions := &vpcclassicv1.GetVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	_, response, err := sess.GetVPNGatewayConnection(getVpnGatewayConnectionOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error getting Vpn Gateway Connection: %s\n%s", err, response)
	}
	return true, nil
}

func vpngwconExists(d *schema.ResourceData, meta interface{}, gID, gConnID string) (bool, error) {
	sess, err := vpcClient(meta)
	if err != nil {
		return false, err
	}

	getVpnGatewayConnectionOptions := &vpcv1.GetVPNGatewayConnectionOptions{
		VPNGatewayID: &gID,
		ID:           &gConnID,
	}
	_, response, err := sess.GetVPNGatewayConnection(getVpnGatewayConnectionOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("Error getting Vpn Gateway Connection: %s\n%s", err, response)
	}
	return true, nil
}
