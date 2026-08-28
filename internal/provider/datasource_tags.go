package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ datasource.DataSource = &TagsDataSource{}

func NewTagsDataSource() datasource.DataSource {
	return &TagsDataSource{}
}

type TagsDataSource struct {
	client *api.Client
}

type TagsDataSourceModel struct {
	ID   types.String   `tfsdk:"id"`
	Tags []TagDataModel `tfsdk:"tags"`
}

type TagDataModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
}

func (d *TagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rms_tags"
}

func (d *TagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of Teltonika RMS Tags.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier for this data source.",
			},
			"tags": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of tags.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.Int64Attribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"color":        schema.StringAttribute{Computed: true},
						"company_id":   schema.Int64Attribute{Computed: true},
						"device_count": schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *TagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TagsDataSourceModel
	var result []map[string]interface{}

	if err := d.client.Get(ctx, "/tags", map[string]string{"limit": "100"}, &result); err != nil {
		resp.Diagnostics.AddError("Error reading tags", fmt.Sprintf("Could not read tags: %s", err))
		return
	}

	var tags []TagDataModel
	for _, t := range result {
		id, ok := t["id"].(float64)
		if !ok {
			resp.Diagnostics.AddError("Error parsing tag ID", "Could not parse id from API response")
			return
		}
		name, ok := t["name"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing tag name", "Could not parse name from API response")
			return
		}
		tag := TagDataModel{
			ID:   types.Int64Value(int64(id)),
			Name: types.StringValue(name),
		}
		if color, ok := t["color"].(string); ok {
			tag.Color = types.StringValue(color)
		}
		if companyID, ok := t["company_id"].(float64); ok {
			tag.CompanyID = types.Int64Value(int64(companyID))
		}
		if deviceCount, ok := t["device_count"].(float64); ok {
			tag.DeviceCount = types.Int64Value(int64(deviceCount))
		}
		tags = append(tags, tag)
	}

	data.ID = types.StringValue("tags-data-source")
	data.Tags = tags
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
