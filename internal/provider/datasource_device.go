package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &DeviceDataSource{}

func NewDeviceDataSource() datasource.DataSource {
	return &DeviceDataSource{}
}

type DeviceDataSource struct {
	client *api.Client
}

type DeviceDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Serial       types.String `tfsdk:"serial"`
	Mac          types.String `tfsdk:"mac"`
	Imei         types.String `tfsdk:"imei"`
	DeviceSeries types.String `tfsdk:"device_series"`
	CompanyID    types.Int64  `tfsdk:"company_id"`
	Status       types.String `tfsdk:"status"`
	Firmware     types.String `tfsdk:"firmware"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (d *DeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *DeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a single Teltonika RMS Device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "The device ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The device name.",
			},
			"serial": schema.StringAttribute{
				Computed:    true,
				Description: "The device serial number.",
			},
			"mac": schema.StringAttribute{
				Computed:    true,
				Description: "The MAC address.",
			},
			"imei": schema.StringAttribute{
				Computed:    true,
				Description: "The IMEI number.",
			},
			"device_series": schema.StringAttribute{
				Computed:    true,
				Description: "The device series.",
			},
			"company_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The company ID.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The device status.",
			},
			"firmware": schema.StringAttribute{
				Computed:    true,
				Description: "The firmware version.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp.",
			},
		},
	}
}

func (d *DeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := d.client.Get(ctx, fmt.Sprintf("/devices/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading device", fmt.Sprintf("Could not read device %d: %s", data.ID.ValueInt64(), err))
		return
	}

	data.Name = types.StringValue(result["name"].(string))
	data.Serial = types.StringValue(result["serial"].(string))
	data.DeviceSeries = types.StringValue(result["device_series"].(string))
	data.Status = types.StringValue(result["status"].(string))

	if mac, ok := result["mac"].(string); ok {
		data.Mac = types.StringValue(mac)
	}
	if imei, ok := result["imei"].(string); ok {
		data.Imei = types.StringValue(imei)
	}
	if firmware, ok := result["firmware"].(string); ok {
		data.Firmware = types.StringValue(firmware)
	}
	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}
	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
