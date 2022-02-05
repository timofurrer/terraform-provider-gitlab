package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/xanzy/go-gitlab"
)

var _ = registerResource("gitlab_project_membership", func() *schema.Resource {
	return &schema.Resource{
		Description: "This resource allows you to add a current user to an existing project with a set access level.",

		CreateContext: resourceGitlabProjectMembershipCreate,
		ReadContext:   resourceGitlabProjectMembershipRead,
		UpdateContext: resourceGitlabProjectMembershipUpdate,
		DeleteContext: resourceGitlabProjectMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "The id of the project.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"user_id": {
				Description: "The id of the user.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Required:    true,
			},
			"access_level": {
				Description:      fmt.Sprintf("The access level for the member. Valid values are: %s", renderValueListForDocs(validProjectAccessLevelNames)),
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validProjectAccessLevelNames, false)),
				Required:         true,
			},
		},
	}
})

func resourceGitlabProjectMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	userId := d.Get("user_id").(int)
	projectId := d.Get("project_id").(string)
	accessLevelId := accessLevelNameToValue[d.Get("access_level").(string)]

	options := &gitlab.AddProjectMemberOptions{
		UserID:      &userId,
		AccessLevel: &accessLevelId,
	}
	log.Printf("[DEBUG] create gitlab project membership for %d in %s", options.UserID, projectId)

	_, _, err := client.ProjectMembers.AddProjectMember(projectId, options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	userIdString := strconv.Itoa(userId)
	d.SetId(buildTwoPartID(&projectId, &userIdString))
	return resourceGitlabProjectMembershipRead(ctx, d, meta)
}

func resourceGitlabProjectMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	id := d.Id()
	log.Printf("[DEBUG] read gitlab project projectMember %s", id)

	projectId, userId, err := projectIdAndUserIdFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	projectMember, resp, err := client.ProjectMembers.GetProjectMember(projectId, userId, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] gitlab project membership for %s not found so removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	resourceGitlabProjectMembershipSetToState(d, projectMember, &projectId)
	return nil
}

func projectIdAndUserIdFromId(id string) (string, int, error) {
	projectId, userIdString, err := parseTwoPartID(id)
	userId, e := strconv.Atoi(userIdString)
	if err != nil {
		e = err
	}
	if e != nil {
		log.Printf("[WARN] cannot get project member id from input: %v", id)
	}
	return projectId, userId, e
}

func resourceGitlabProjectMembershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	userId := d.Get("user_id").(int)
	projectId := d.Get("project_id").(string)
	accessLevelId := accessLevelNameToValue[strings.ToLower(d.Get("access_level").(string))]

	options := gitlab.EditProjectMemberOptions{
		AccessLevel: &accessLevelId,
	}
	log.Printf("[DEBUG] update gitlab project membership %v for %s", userId, projectId)

	_, _, err := client.ProjectMembers.EditProjectMember(projectId, userId, &options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGitlabProjectMembershipRead(ctx, d, meta)
}

func resourceGitlabProjectMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	id := d.Id()
	projectId, userId, err := projectIdAndUserIdFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Delete gitlab project membership %v for %s", userId, projectId)

	_, err = client.ProjectMembers.DeleteProjectMember(projectId, userId, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGitlabProjectMembershipSetToState(d *schema.ResourceData, projectMember *gitlab.ProjectMember, projectId *string) {

	d.Set("project_id", projectId)
	d.Set("user_id", projectMember.ID)
	d.Set("access_level", accessLevelValueToName[projectMember.AccessLevel])

	userId := strconv.Itoa(projectMember.ID)
	d.SetId(buildTwoPartID(projectId, &userId))
}
