package buildkite

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type organizationResourceModel struct {
	AllowedApiIpAddresses types.List   `tfsdk:"allowed_api_ip_addresses"`
	ID                    types.String `tfsdk:"id"`
	UUID                  types.String `tfsdk:"uuid"`
}

type organizationResource struct {
	client *Client
}

func newOrganizationResource() resource.Resource {
	return &organizationResource{}
}

func (organizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (o *organizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	o.client = req.ProviderData.(*Client)
}

func (*organizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_schema.Schema{
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": resource_schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_api_ip_addresses": resource_schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (o *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, state organizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create CIDR slice from AllowedApiIpAddresses
	cidrs := createCidrSliceFromList(plan.AllowedApiIpAddresses)

	log.Printf("Creating settings for organization %s ...", o.client.organizationId)
	apiResponse, err := setApiIpAddresses(
		o.client.genqlient,
		o.client.organizationId,
		strings.Join(cidrs, " "),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Organization settings",
			fmt.Sprintf("Unable to create Organization settings: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(apiResponse.OrganizationApiIpAllowlistUpdate.Organization.Id)
	state.UUID = types.StringValue(apiResponse.OrganizationApiIpAllowlistUpdate.Organization.Uuid)
	ips, diag := types.ListValueFrom(ctx, types.StringType, strings.Split(apiResponse.OrganizationApiIpAllowlistUpdate.Organization.AllowedApiIpAddresses, " "))
	state.AllowedApiIpAddresses = ips

	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (o *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("Reading settings for organization ...")
	response, err := getOrganization(o.client.genqlient, o.client.organization)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to obtain Organization",
			fmt.Sprintf("Unable to obtain Organization: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(o.client.organizationId)
	state.UUID = types.StringValue(response.Organization.Uuid)
	ips, diag := types.ListValueFrom(ctx, types.StringType, strings.Split(response.Organization.AllowedApiIpAddresses, " "))
	state.AllowedApiIpAddresses = ips

	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (o *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (o *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state organizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create CIDR slice from AllowedApiIpAddresses
	cidrs := createCidrSliceFromList(plan.AllowedApiIpAddresses)

	log.Printf("Updating settings for organization %s ...", o.client.organizationId)
	apiResponse, err := setApiIpAddresses(
		o.client.genqlient,
		o.client.organizationId,
		strings.Join(cidrs, " "),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update Organization settings",
			fmt.Sprintf("Unable to update Organization settings: %s", err.Error()),
		)
		return
	}
	state.ID = types.StringValue(apiResponse.OrganizationApiIpAllowlistUpdate.Organization.Id)
	state.UUID = types.StringValue(apiResponse.OrganizationApiIpAllowlistUpdate.Organization.Uuid)
	ips, diag := types.ListValueFrom(ctx, types.StringType, strings.Split(apiResponse.OrganizationApiIpAllowlistUpdate.Organization.AllowedApiIpAddresses, " "))
	state.AllowedApiIpAddresses = ips

	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (o *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	log.Printf("Deleting settings for organization %s ...", o.client.organizationId)

	_, err := setApiIpAddresses(o.client.genqlient, o.client.organizationId, "")

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Organization settings",
			fmt.Sprintf("Unable to delete Organization settings: %s", err.Error()),
		)
		return
	}
}

func createCidrSliceFromList(cidrList types.List) []string {
	cidrs := make([]string, len(cidrList.Elements()))
	for i, v := range cidrList.Elements() {
		cidrs[i] = strings.Trim(v.String(), "\"")
	}

	return cidrs
}