package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// DynamicValue returns the fwschemadata.Data for a given
// *tfprotov5.DynamicValue.
//
// If necessary, the underlying data is modified to convert list and set block
// values from an empty collection to a null collection. This is to prevent
// developers from needing to understand Terraform's differences between
// block and attribute values where blocks are technically never null, but from
// a developer perspective this distinction introduces unnecessary complexity.
// This null block value translation is automatically handled by Terraform from
// provider responses.
func DynamicValue(ctx context.Context, proto5 *tfprotov5.DynamicValue, schema fwschema.Schema, description fwschemadata.DataDescription) (fwschemadata.Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	data := &fwschemadata.Data{
		Description: description,
		Schema:      schema,
	}

	if proto5 == nil {
		return *data, diags
	}

	proto5Value, err := proto5.Unmarshal(schema.Type().TerraformType(ctx))

	if err != nil {
		diags.AddError(
			"Unable to Convert "+description.Title(),
			"An unexpected error was encountered when converting the "+description.String()+" from the protocol type. "+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Unable to unmarshal DynamicValue: "+err.Error(),
		)

		return *data, diags
	}

	data.TerraformValue = proto5Value

	diags.Append(data.NullifyCollectionBlocks(ctx)...)

	return *data, diags
}
