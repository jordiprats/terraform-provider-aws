package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	userStateExists = "exists"
)

func ServerState(conn *transfer.Transfer, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ServerByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func UserState(conn *transfer.Transfer, serverID, userName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.UserByServerIDAndUserName(conn, serverID, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, userStateExists, nil
	}
}
