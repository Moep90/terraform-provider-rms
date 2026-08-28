package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &DevicesDataSource{}

func NewDevicesDataSource() datasource.DataSource {
	return &DevicesDataSource{}
}

type DevicesDataSource struct {
	client *api.Client
}

type DevicesDataSourceModel struct {
	ID        types.String      `tfsdk:"id"`
	Devices   []DeviceDataModel `tfsdk:"devices"`
	CompanyID types.Int64       `tfsdk:"company_id"`
	Status    types.String      `tfsdk:"status"`
}

type DeviceDataModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Serial       types.String `tfsdk:"serial"`
	Mac          types.String `tfsdk:"mac"`
	Imei         types.String `tfsdk:"imei"`
	DeviceSeries types.String `tfsdk:"device_series"`
	Status       types.String `tfsdk:"status"`
	Firmware     types.String `tfsdk:"firmware"`
}

func (d *DevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_devices"
}

func (d *DevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of Teltonika RMS Devices.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"company_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Filter by company ID.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by status: online, offline, not_activated.",
			},
			"devices": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of devices.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.Int64Attribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"serial":        schema.StringAttribute{Computed: true},
						"mac":           schema.StringAttribute{Computed: true},
						"imei":          schema.StringAttribute{Computed: true},
						"device_series": schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"firmware":      schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *DevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := map[string]string{"limit": "100"}
	if !data.CompanyID.IsNull() {
		params["company_id"] = fmt.Sprintf("%d", data.CompanyID.ValueInt64())
	}
	if !data.Status.IsNull() {
		params["status"] = data.Status.ValueString()
	}

	var result []map[string]interface{}
	if err := d.client.Get(ctx, "/devices", params, &result); err != nil {
		resp.Diagnostics.AddError("Error reading devices", fmt.Sprintf("Could not read devices: %s", err))
		return
	}

	var devices []DeviceDataModel
	for _, dev := range result {
		id, ok := dev["id"].(float64)
		if !ok {
			resp.Diagnostics.AddError("Error parsing device ID", "Could not parse id from API response")
			return
		}
		name, ok := dev["name"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing device name", "Could not parse name from API response")
			return
		}
		serial, ok := dev["serial"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing device serial", "Could not parse serial from API response")
			return
		}
		deviceSeries, ok := dev["device_series"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing device series", "Could not parse device_series from API response")
			return
		}
		status, ok := dev["status"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing device status", "Could not parse status from API response")
			return
		}
		device := DeviceDataModel{
			ID:           types.Int64Value(int64(id)),
			Name:         types.StringValue(name),
			Serial:       types.StringValue(serial),
			DeviceSeries: types.StringValue(deviceSeries),
			Status:       types.StringValue(status),
		}
		if mac, ok := dev["mac"].(string); ok {
			device.Mac = types.StringValue(mac)
		}
		if imei, ok := dev["imei"].(string); ok {
			device.Imei = types.StringValue(imei)
		}
		if firmware, ok := dev["firmware"].(string); ok {
			device.Firmware = types.StringValue(firmware)
		}
		devices = append(devices, device)
	}

	data.ID = types.StringValue("devices-data-source")
	data.Devices = devices
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
