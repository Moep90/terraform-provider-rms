package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &DeviceEsimBootstrapDataSource{}

func NewDeviceEsimBootstrapDataSource() datasource.DataSource {
	return &DeviceEsimBootstrapDataSource{}
}

type DeviceEsimBootstrapDataSource struct {
	client *api.Client
}

type DeviceEsimBootstrapDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	DeviceID      types.Int64  `tfsdk:"device_id"`
	EsimBootstrap types.String `tfsdk:"esim_bootstrap"`
	Status        types.String `tfsdk:"status"`
	Message       types.String `tfsdk:"message"`
}

func (d *DeviceEsimBootstrapDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_device_esim_bootstrap"
}

func (d *DeviceEsimBootstrapDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the E-SIM bootstrap status for a Teltonika RMS device. Useful for TRB devices with eSIM capabilities.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"device_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the device to check E-SIM bootstrap status for.",
			},
			"esim_bootstrap": schema.StringAttribute{
				Computed:    true,
				Description: "The E-SIM bootstrap status (e.g., 'enabled', 'disabled', 'pending').",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The overall status of the E-SIM bootstrap check.",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Additional message about the E-SIM bootstrap status.",
			},
		},
	}
}

func (d *DeviceEsimBootstrapDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *DeviceEsimBootstrapDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceEsimBootstrapDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	path := fmt.Sprintf("/devices/%d/check/esim-bootstrap", data.DeviceID.ValueInt64())
	if err := d.client.Get(ctx, path, nil, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error reading E-SIM bootstrap status",
			fmt.Sprintf("Could not read E-SIM bootstrap status for device %d: %s", data.DeviceID.ValueInt64(), err),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("device-%d-esim-bootstrap", data.DeviceID.ValueInt64()))

	if esimBootstrap, ok := result["esim_bootstrap"].(string); ok {
		data.EsimBootstrap = types.StringValue(esimBootstrap)
	}

	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}

	if message, ok := result["message"].(string); ok {
		data.Message = types.StringValue(message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
