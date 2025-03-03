package qldb

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qldb_ledger", name="Ledger")
// @Tags(identifierAttribute="arn")
func ResourceLedger() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLedgerCreate,
		ReadWithoutTimeout:   resourceLedgerRead,
		UpdateWithoutTimeout: resourceLedgerUpdate,
		DeleteWithoutTimeout: resourceLedgerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"kms_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"AWS_OWNED_KMS_KEY"}, false),
					verify.ValidARN,
				),
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"permissions_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(qldb.PermissionsMode_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLedgerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()

	name := create.Name(d.Get("name").(string), "tf")
	input := &qldb.CreateLedgerInput{
		DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
		Name:               aws.String(name),
		PermissionsMode:    aws.String(d.Get("permissions_mode").(string)),
		Tags:               GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("kms_key"); ok {
		input.KmsKey = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating QLDB Ledger: %s", input)
	output, err := conn.CreateLedgerWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating QLDB Ledger (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Name))

	if _, err := waitLedgerCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return diag.Errorf("waiting for QLDB Ledger (%s) create: %s", d.Id(), err)
	}

	return resourceLedgerRead(ctx, d, meta)
}

func resourceLedgerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()

	ledger, err := FindLedgerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QLDB Ledger %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading QLDB Ledger (%s): %s", d.Id(), err)
	}

	d.Set("arn", ledger.Arn)
	d.Set("deletion_protection", ledger.DeletionProtection)
	if ledger.EncryptionDescription != nil {
		d.Set("kms_key", ledger.EncryptionDescription.KmsKeyArn)
	} else {
		d.Set("kms_key", nil)
	}
	d.Set("name", ledger.Name)
	d.Set("permissions_mode", ledger.PermissionsMode)

	return nil
}

func resourceLedgerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()

	if d.HasChange("permissions_mode") {
		input := &qldb.UpdateLedgerPermissionsModeInput{
			Name:            aws.String(d.Id()),
			PermissionsMode: aws.String(d.Get("permissions_mode").(string)),
		}

		log.Printf("[INFO] Updating QLDB Ledger permissions mode: %s", input)
		if _, err := conn.UpdateLedgerPermissionsModeWithContext(ctx, input); err != nil {
			return diag.Errorf("updating QLDB Ledger (%s) permissions mode: %s", d.Id(), err)
		}
	}

	if d.HasChanges("deletion_protection", "kms_key") {
		input := &qldb.UpdateLedgerInput{
			DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
			Name:               aws.String(d.Id()),
		}

		if d.HasChange("kms_key") {
			input.KmsKey = aws.String(d.Get("kms_key").(string))
		}

		log.Printf("[INFO] Updating QLDB Ledger: %s", input)
		if _, err := conn.UpdateLedgerWithContext(ctx, input); err != nil {
			return diag.Errorf("updating QLDB Ledger (%s): %s", d.Id(), err)
		}
	}

	return resourceLedgerRead(ctx, d, meta)
}

func resourceLedgerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBConn()

	input := &qldb.DeleteLedgerInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting QLDB Ledger: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute,
		func() (interface{}, error) {
			return conn.DeleteLedgerWithContext(ctx, input)
		}, qldb.ErrCodeResourceInUseException)

	if tfawserr.ErrCodeEquals(err, qldb.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting QLDB Ledger (%s): %s", d.Id(), err)
	}

	if _, err := waitLedgerDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete), d.Id()); err != nil {
		return diag.Errorf("waiting for QLDB Ledger (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindLedgerByName(ctx context.Context, conn *qldb.QLDB, name string) (*qldb.DescribeLedgerOutput, error) {
	input := &qldb.DescribeLedgerInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeLedgerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, qldb.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if state := aws.StringValue(output.State); state == qldb.LedgerStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusLedgerState(ctx context.Context, conn *qldb.QLDB, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLedgerByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitLedgerCreated(ctx context.Context, conn *qldb.QLDB, timeout time.Duration, name string) (*qldb.DescribeLedgerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{qldb.LedgerStateCreating},
		Target:     []string{qldb.LedgerStateActive},
		Refresh:    statusLedgerState(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qldb.DescribeLedgerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitLedgerDeleted(ctx context.Context, conn *qldb.QLDB, timeout time.Duration, name string) (*qldb.DescribeLedgerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{qldb.LedgerStateActive, qldb.LedgerStateDeleting},
		Target:     []string{},
		Refresh:    statusLedgerState(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qldb.DescribeLedgerOutput); ok {
		return output, err
	}

	return nil, err
}
