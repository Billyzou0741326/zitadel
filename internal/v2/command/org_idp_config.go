package command

import (
	"context"
	"github.com/caos/zitadel/internal/eventstore/v2"

	caos_errs "github.com/caos/zitadel/internal/errors"
	"github.com/caos/zitadel/internal/telemetry/tracing"
	"github.com/caos/zitadel/internal/v2/domain"

	"github.com/caos/zitadel/internal/crypto"
	"github.com/caos/zitadel/internal/errors"
	org_repo "github.com/caos/zitadel/internal/v2/repository/org"
)

func (r *CommandSide) AddIDPConfig(ctx context.Context, config *domain.IDPConfig) (*domain.IDPConfig, error) {
	if config.OIDCConfig == nil {
		return nil, errors.ThrowInvalidArgument(nil, "Org-eUpQU", "Errors.idp.config.notset")
	}

	idpConfigID, err := r.idGenerator.Next()
	if err != nil {
		return nil, err
	}
	addedConfig := NewOrgIDPConfigWriteModel(idpConfigID, config.AggregateID)

	clientSecret, err := crypto.Crypt([]byte(config.OIDCConfig.ClientSecretString), r.idpConfigSecretCrypto)
	if err != nil {
		return nil, err
	}

	orgAgg := OrgAggregateFromWriteModel(&addedConfig.WriteModel)
	events := []eventstore.EventPusher{
		org_repo.NewIDPConfigAddedEvent(
			ctx,
			orgAgg,
			idpConfigID,
			config.Name,
			config.Type,
			config.StylingType,
		),
		org_repo.NewIDPOIDCConfigAddedEvent(
			ctx,
			orgAgg,
			config.OIDCConfig.ClientID,
			idpConfigID,
			config.OIDCConfig.Issuer,
			clientSecret,
			config.OIDCConfig.IDPDisplayNameMapping,
			config.OIDCConfig.UsernameMapping,
			config.OIDCConfig.Scopes...),
	}
	pushedEvents, err := r.eventstore.PushEvents(ctx, events...)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(addedConfig, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToIDPConfig(&addedConfig.IDPConfigWriteModel), nil
}

func (r *CommandSide) ChangeIDPConfig(ctx context.Context, config *domain.IDPConfig) (*domain.IDPConfig, error) {
	existingIDP, err := r.orgIDPConfigWriteModelByID(ctx, config.IDPConfigID, config.AggregateID)
	if err != nil {
		return nil, err
	}
	if existingIDP.State == domain.IDPConfigStateRemoved || existingIDP.State == domain.IDPConfigStateUnspecified {
		return nil, caos_errs.ThrowNotFound(nil, "Org-4M9so", "Errors.Org.IDPConfig.NotExisting")
	}

	orgAgg := OrgAggregateFromWriteModel(&existingIDP.WriteModel)
	changedEvent, hasChanged := existingIDP.NewChangedEvent(
		ctx,
		orgAgg,
		config.IDPConfigID,
		config.Name,
		config.StylingType)

	if !hasChanged {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "Org-4M9vs", "Errors.Org.LabelPolicy.NotChanged")
	}
	pushedEvents, err := r.eventstore.PushEvents(ctx, changedEvent)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingIDP, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToIDPConfig(&existingIDP.IDPConfigWriteModel), nil
}

func (r *CommandSide) DeactivateIDPConfig(ctx context.Context, idpID, orgID string) error {
	existingIDP, err := r.orgIDPConfigWriteModelByID(ctx, idpID, orgID)
	if err != nil {
		return err
	}
	if existingIDP.State != domain.IDPConfigStateActive {
		return caos_errs.ThrowPreconditionFailed(nil, "Org-4M9so", "Errors.Org.IDPConfig.NotActive")
	}
	orgAgg := OrgAggregateFromWriteModel(&existingIDP.WriteModel)
	_, err = r.eventstore.PushEvents(ctx, org_repo.NewIDPConfigDeactivatedEvent(ctx, orgAgg, idpID))
	return err
}

func (r *CommandSide) ReactivateIDPConfig(ctx context.Context, idpID, orgID string) error {
	existingIDP, err := r.orgIDPConfigWriteModelByID(ctx, idpID, orgID)
	if err != nil {
		return err
	}
	if existingIDP.State != domain.IDPConfigStateInactive {
		return caos_errs.ThrowPreconditionFailed(nil, "Org-5Mo0d", "Errors.Org.IDPConfig.NotInactive")
	}
	orgAgg := OrgAggregateFromWriteModel(&existingIDP.WriteModel)
	_, err = r.eventstore.PushEvents(ctx, org_repo.NewIDPConfigReactivatedEvent(ctx, orgAgg, idpID))
	return err
}

func (r *CommandSide) RemoveIDPConfig(ctx context.Context, idpID, orgID string, cascadeRemoveProvider bool, cascadeExternalIDPs ...*domain.ExternalIDP) error {
	existingIDP, err := r.orgIDPConfigWriteModelByID(ctx, idpID, orgID)
	if err != nil {
		return err
	}

	if existingIDP.State == domain.IDPConfigStateRemoved || existingIDP.State == domain.IDPConfigStateUnspecified {
		return caos_errs.ThrowNotFound(nil, "Org-Yx9vd", "Errors.Org.IDPConfig.NotExisting")
	}
	if existingIDP.State != domain.IDPConfigStateInactive {
		return caos_errs.ThrowPreconditionFailed(nil, "Org-5Mo0d", "Errors.Org.IDPConfig.NotInactive")
	}

	orgAgg := OrgAggregateFromWriteModel(&existingIDP.WriteModel)
	events := []eventstore.EventPusher{
		org_repo.NewIDPConfigRemovedEvent(ctx, orgAgg, idpID, existingIDP.Name),
	}

	if cascadeRemoveProvider {
		removeIDPEvents := r.removeIDPProviderFromLoginPolicy(ctx, orgAgg, idpID, true, cascadeExternalIDPs...)
		events = append(events, removeIDPEvents...)
	}
	_, err = r.eventstore.PushEvents(ctx, events...)
	return err
}

func (r *CommandSide) orgIDPConfigWriteModelByID(ctx context.Context, idpID, orgID string) (policy *OrgIDPConfigWriteModel, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel := NewOrgIDPConfigWriteModel(idpID, orgID)
	err = r.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}