package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &DevicesExportDataSource{}

func NewDevicesExportDataSource() datasource.DataSource {
	return &DevicesExportDataSource{}
}

type DevicesExportDataSource struct {
	client *api.Client
}

type DevicesExportDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	CSVData types.String `tfsdk:"csv_data"`
}

func (d *DevicesExportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_devices_export"
}

func (d *DevicesExportDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Exports device information in CSV format from Teltonika RMS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"csv_data": schema.StringAttribute{
				Computed:    true,
				Description: "The raw CSV data containing device information.",
			},
		},
	}
}

func (d *DevicesExportDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DevicesExportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DevicesExportDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.GetRaw(ctx, "/devices/export/csv", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching CSV export",
			fmt.Sprintf("Could not fetch CSV export: %s", err),
		)
		return
	}

	data.ID = types.StringValue("devices-export-csv")
	data.CSVData = types.StringValue(string(body))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
