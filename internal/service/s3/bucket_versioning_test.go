package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketVersioning_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningBasicConfig(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketVersioning_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningBasicConfig(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketVersioning(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketVersioning_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningBasicConfig(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
				),
			},
			{
				Config: testAccBucketVersioningBasicConfig(rName, s3.BucketVersioningStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusSuspended),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketVersioningBasicConfig(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
		},
	})
}

// TestAccBucketVersioning_MFADelete can only test for a "Disabled"
// mfa_delete configuration as the "mfa" argument is required if it's enabled
func TestAccS3BucketVersioning_MFADelete(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_MFADelete(rName, s3.MFADeleteDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.mfa_delete", s3.MFADeleteDisabled),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBucketVersioningDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_versioning" {
			continue
		}

		input := &s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBucketVersioning(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 bucket versioning (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && aws.StringValue(output.Status) != s3.BucketVersioningStatusSuspended {
			return fmt.Errorf("S3 bucket versioning (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketVersioningExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		input := &s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBucketVersioning(input)

		if err != nil {
			return fmt.Errorf("error getting S3 bucket versioning (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("S3 Bucket versioning (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketVersioningBasicConfig(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = %[2]q
  }
}
`, rName, status)
}

func testAccBucketVersioningConfig_MFADelete(rName, mfaDelete string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    mfa_delete = %[2]q
    status     = "Enabled"
  }
}
`, rName, mfaDelete)
}
