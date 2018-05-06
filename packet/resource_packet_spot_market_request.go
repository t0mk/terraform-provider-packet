package packet

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/packethost/packngo"
)

func resourcePacketSpotMarketRequest() *schema.Resource {
	return &schema.Resource{
		Create: resourcePacketSpotMarketRequestCreate,
		Read:   resourcePacketSpotMarketRequestRead,
		Delete: resourcePacketSpotMarketRequestDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"facilities": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"devices_max": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"devices_min": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"devices": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"end_at": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"billing_cycle": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"updated": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePacketSpotMarketRequestCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	createRequest := &packngo.SpotMarketRequestCreateRequest{
		PlanID:     d.Get("plan").(string),
		FacilityID: d.Get("facility").(string),
		Size:       d.Get("size").(int),
		Locked:     d.Get("locked").(bool),
	}

	if attr, ok := d.GetOk("billing_cycle"); ok {
		createRequest.BillingCycle = attr.(string)
	} else {
		createRequest.BillingCycle = "hourly"
	}

	if attr, ok := d.GetOk("description"); ok {
		createRequest.Description = attr.(string)
	}

	snapshot_count := d.Get("snapshot_policies.#").(int)
	if snapshot_count > 0 {
		createRequest.SnapshotPolicies = make([]*packngo.SnapshotPolicy, 0, snapshot_count)
		for i := 0; i < snapshot_count; i++ {
			policy := new(packngo.SnapshotPolicy)
			policy.SnapshotFrequency = d.Get(fmt.Sprintf("snapshot_policies.%d.snapshot_frequency", i)).(string)
			policy.SnapshotCount = d.Get(fmt.Sprintf("snapshot_policies.%d.snapshot_count", i)).(int)
			createRequest.SnapshotPolicies = append(createRequest.SnapshotPolicies, policy)
		}
	}

	newSpotMarketRequest, _, err := client.SpotMarketRequests.Create(createRequest, d.Get("project_id").(string))
	if err != nil {
		return friendlyError(err)
	}

	d.SetId(newSpotMarketRequest.ID)

	_, err = waitForSpotMarketRequestAttribute(d, "active", []string{"queued", "provisioning"}, "state", meta)
	if err != nil {
		if isForbidden(err) {
			// If the volume doesn't get to the active state, we can't recover it from here.
			d.SetId("")

			return errors.New("provisioning time limit exceeded; the Packet team will investigate")
		}
		return err
	}

	return resourcePacketSpotMarketRequestRead(d, meta)
}

func waitForSpotMarketRequestAttribute(d *schema.ResourceData, target string, pending []string, attribute string, meta interface{}) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     []string{target},
		Refresh:    newSpotMarketRequestStateRefreshFunc(d, attribute, meta),
		Timeout:    60 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	return stateConf.WaitForState()
}

func newSpotMarketRequestStateRefreshFunc(d *schema.ResourceData, attribute string, meta interface{}) resource.StateRefreshFunc {
	client := meta.(*packngo.Client)

	return func() (interface{}, string, error) {
		if err := resourcePacketSpotMarketRequestRead(d, meta); err != nil {
			return nil, "", err
		}

		if attr, ok := d.GetOk(attribute); ok {
			volume, _, err := client.SpotMarketRequests.Get(d.Id())
			if err != nil {
				return nil, "", friendlyError(err)
			}
			return &volume, attr.(string), nil
		}

		return nil, "", nil
	}
}

func resourcePacketSpotMarketRequestRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	volume, _, err := client.SpotMarketRequests.Get(d.Id())
	if err != nil {
		err = friendlyError(err)

		// If the volume somehow already destroyed, mark as succesfully gone.
		if isNotFound(err) {
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", volume.Name)
	d.Set("description", volume.Description)
	d.Set("size", volume.Size)
	d.Set("plan", volume.Plan.Slug)
	d.Set("facility", volume.Facility.Code)
	d.Set("state", volume.State)
	d.Set("billing_cycle", volume.BillingCycle)
	d.Set("locked", volume.Locked)
	d.Set("created", volume.Created)
	d.Set("updated", volume.Updated)

	snapshot_policies := make([]map[string]interface{}, 0, len(volume.SnapshotPolicies))
	for _, snapshot_policy := range volume.SnapshotPolicies {
		policy := map[string]interface{}{
			"snapshot_frequency": snapshot_policy.SnapshotFrequency,
			"snapshot_count":     snapshot_policy.SnapshotCount,
		}
		snapshot_policies = append(snapshot_policies, policy)
	}
	d.Set("snapshot_policies", snapshot_policies)

	attachments := make([]*packngo.SpotMarketRequestAttachment, 0, len(volume.Attachments))
	for _, attachment := range volume.Attachments {
		attachments = append(attachments, attachment)
	}
	d.Set("attachments", attachments)

	return nil
}

func resourcePacketSpotMarketRequestUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	if d.HasChange("locked") {
		// the change is true => false, i.e. unlock
		if !d.Get("locked").(bool) {
			if _, err := client.SpotMarketRequests.Unlock(d.Id()); err != nil {
				return friendlyError(err)
			}
		}
	}

	updateRequest := &packngo.SpotMarketRequestUpdateRequest{}

	sendAttrUpdate := false

	if d.HasChange("description") {
		sendAttrUpdate = true
		vDesc := d.Get("description").(string)
		updateRequest.Description = &vDesc
	}
	if d.HasChange("plan") {
		sendAttrUpdate = true
		vPlan := d.Get("plan").(string)
		updateRequest.PlanID = &vPlan
	}
	if d.HasChange("size") {
		sendAttrUpdate = true
		vSize := d.Get("size").(int)
		updateRequest.Size = &vSize
	}
	if d.HasChange("billing_cycle") {
		sendAttrUpdate = true
		vCycle := d.Get("billing_cycle").(string)
		updateRequest.BillingCycle = &vCycle
	}

	if sendAttrUpdate {
		_, _, err := client.SpotMarketRequests.Update(d.Id(), updateRequest)
		if err != nil {
			return friendlyError(err)
		}
	}
	if d.HasChange("locked") {
		// the change is false => true, i.e. lock
		if d.Get("locked").(bool) {
			if _, err := client.SpotMarketRequests.Lock(d.Id()); err != nil {
				return friendlyError(err)
			}
		}
	}

	return resourcePacketSpotMarketRequestRead(d, meta)
}

func resourcePacketSpotMarketRequestDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*packngo.Client)

	if _, err := client.SpotMarketRequests.Delete(d.Id()); err != nil {
		return friendlyError(err)
	}

	return nil
}
