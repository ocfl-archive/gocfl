package inventoryimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"golang.org/x/exp/slices"
)

type versionValue map[*inventory.VersionNumber]uint

func (v versionValue) GetInt(version *inventory.VersionNumber) (uint, bool) {
	vInt, ok := v[version]
	return vInt, ok
}

func (v versionValue) GetVersion(version uint) (*inventory.VersionNumber, bool) {
	for ver, verInt := range v {
		if verInt == version {
			return ver, true
		}
	}
	return nil, false
}

func NewVersionsBase(ctx context.Context, factory inventory.Factory, logger ocfllogger.OCFLLogger) inventory.Versions {
	return &versionsBase{
		ctx:         ctx,
		versions:    map[int]inventory.Version{},
		versionInts: map[string]int{},
		factory:     factory,
		logger:      logger,
	}
}

type versionsBase struct {
	versions    map[int]inventory.Version
	versionInts map[string]int
	err         error
	//paddingLength int
	factory inventory.Factory
	logger  ocfllogger.OCFLLogger
	ctx     context.Context
}

// todo: remove head parameter
func (v *versionsBase) NewVersion(msg string, UserName string, UserAddress string) error {
	var newVersionNumber *inventory.VersionNumber
	head := v.LatestVersionNumber()
	if head.IsValid() {
		if paddingLength := head.GetPaddingLength(); paddingLength > 0 {
			newVersionNumber = inventory.NewVersionNumber().WithString(fmt.Sprintf("v0%0*d", paddingLength, head.Int()+1))
		} else {
			newVersionNumber = inventory.NewVersionNumber().WithString(fmt.Sprintf("v%d", head.Int()+1))
		}
	} else {
		newVersionNumber = inventory.NewVersionNumber().WithString("v1")
	}
	state := v.factory.NewState(v.ctx)
	if head.IsValid() {
		if err := state.CopyFrom(v.GetVersion(head).GetState()); err != nil {
			return errors.Wrapf(err, "Failed to copy version from %s", head)
		}
	}
	user := v.factory.NewUser(v.ctx).WithAddress(UserAddress).WithName(UserName)
	ver := v.factory.NewVersion(v.ctx).
		WithCreated(time.Now()).
		WithMessage(msg).
		WithState(state).
		WithUser(user).
		WithVersion(newVersionNumber)
	v.SetVersion(newVersionNumber, ver)
	return nil
}

func (v *versionsBase) Err() error {
	if v == nil {
		return nil
	}
	var errs = []error{}
	for _, version := range v.Iterate() {
		errs = append(errs, version.Err())
	}
	return errors.Combine(errs...)
}

func (v *versionsBase) Delete(versionNumber *inventory.VersionNumber) (bool, error) {
	if _, ok := v.versions[versionNumber.Int()]; !ok {
		return false, nil
	}
	delete(v.versions, versionNumber.Int())
	delete(v.versionInts, versionNumber.String())
	return true, nil
}

func (v *versionsBase) AddFile(stateFilename string, digest string) (bool, error) {
	latestVersionString := v.LatestVersionNumber()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.AddFile(stateFilename, digest)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to add file %s [%s]", stateFilename, digest)
	}
	return modified, nil
}

func (v *versionsBase) EchoDelete(existing []string, pathPrefix string) (bool, error) {
	latestVersionString := v.LatestVersionNumber()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.EchoDelete(existing, pathPrefix)
	if err != nil {
		return modified, errors.WithStack(err)
	}
	return modified, nil
}

func (v *versionsBase) CopyFile(stateFilename, digest string) (bool, error) {
	latestVersionString := v.LatestVersionNumber()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.CopyFile(stateFilename, digest)
	if err != nil {
		return false, errors.Wrapf(err, "failed to copy file with digest '%s' to '%s' from version %s", digest, stateFilename, latestVersionString)
	}
	return modified, nil
}

func (v *versionsBase) RenameFile(oldStateFilename, newStateFilename string) (bool, error) {
	latestVersionString := v.LatestVersionNumber()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.RenameFile(oldStateFilename, newStateFilename)
	if err != nil {
		return false, errors.Wrapf(err, "failed to rename file '%s' to '%s' from version %s", oldStateFilename, newStateFilename, latestVersionString)
	}
	return modified, nil
}

func (v *versionsBase) DeleteFile(stateFilename string) (bool, error) {
	latestVersionString := v.LatestVersionNumber()
	latestVersion := v.GetVersion(latestVersionString)
	if latestVersion == nil {
		return false, errors.Errorf("version %s not found", latestVersionString)
	}
	modified, err := latestVersion.DeleteFile(stateFilename)
	if err != nil {
		return false, errors.Wrapf(err, "failed to delete file %s from version %s", stateFilename, latestVersionString)
	}
	return modified, nil
}

func (v *versionsBase) FileExists(path, digest string) (bool, error) {
	// find all versions, which contain the path
	css := map[int]string{}
	for versionString, ver := range v.Iterate() {
		cs := ver.FileChecksum(path)
		if cs == "" {
			continue
		}
		css[versionString.Int()] = cs
	}
	if len(css) == 0 {
		return false, nil
	}
	latestVersion := v.LatestVersionNumber()

	lastChecksum, ok := css[latestVersion.Int()]
	if !ok {
		return false, errors.Errorf("checksum for version %s does not exist", lastChecksum)
	}

	// check whether the latest version is the correct file
	return digest == "" || lastChecksum == digest, nil
}

func (v *versionsBase) GetVersionNumbers() iter.Seq[*inventory.VersionNumber] {
	return func(yield func(*inventory.VersionNumber) bool) {
		for versionString, versionInt := range v.versionInts {
			if !yield(inventory.NewVersionNumber().Init(versionInt, versionString)) {
				return
			}
		}
	}
}

func (v *versionsBase) Finalize(inCreation bool) error {
	for ver, version := range v.Iterate() {
		if err := version.Finalize(inCreation); err != nil {
			return errors.Wrapf(err, "failed to finalize inventory version '%s'", ver)
		}
	}
	return nil
}

func (v *versionsBase) Check(manifestDigests []string) error {
	if v.IsEmpty() {
		v.logger.ValidationError(validation.E008, "length of ver is 0")
		return nil
	}
	var versionsSeq = []int{}
	var paddingLength int = -1
	manifestDigestsLower := []string{}
	for _, manifestDigest := range manifestDigests {
		manifestDigestsLower = append(manifestDigestsLower, strings.ToLower(manifestDigest))
	}
	slices.Sort(manifestDigests)
	slices.Sort(manifestDigestsLower)
	for versionNumber, ver := range v.Iterate() {
		versionsSeq = append(versionsSeq, versionNumber.Int())
		if inventory.VersionZeroRegexp.MatchString(versionNumber.String()) {
			if paddingLength == -1 {
				paddingLength = len(versionNumber.String()) - 2
			} else {
				if paddingLength != len(versionNumber.String())-2 {
					//i.AddValidationError(E011, "invalid version padding '%s'", ver)
					v.logger.ValidationError(validation.E012, "invalid version padding %s", versionNumber.String())
					v.logger.ValidationError(validation.E013, "invalid version padding %s", versionNumber.String())
				}
			}
		} else {
			if inventory.VersionNoZeroRegexp.MatchString(versionNumber.String()) {
				if paddingLength == -1 {
					paddingLength = 0
				} else {
					if paddingLength != 0 {
						v.logger.ValidationError(validation.E011, "invalid version padding %s", versionNumber.String())
						v.logger.ValidationError(validation.E012, "invalid version padding %s", versionNumber.String())
						v.logger.ValidationError(validation.E013, "invalid version padding %s", versionNumber.String())
					}
				}
			} else {
				// todo: this error is only for ocfl 1.1, find solution for ocfl 1.0
				v.logger.ValidationError(validation.E014, "invalid version format %s", versionNumber.String())
				v.logger.ValidationError(validation.E001, "invalid version format %s", versionNumber.String())
			}
		}

		if err := ver.Check(manifestDigests, manifestDigestsLower); err != nil {
			return errors.Wrapf(err, "version %s validation check failed", versionNumber)
		}
	}
	slices.Sort(versionsSeq)
	for key, vInt := range versionsSeq {
		if key != vInt-1 {
			v.logger.ValidationError(validation.E010, "invalid version sequence %v", versionsSeq)
			break
		}
	}
	if paddingLength > 0 {
		v.logger.ValidationError(validation.W001, "padding length %d", paddingLength)
	}

	return nil
}

func (v *versionsBase) SetVersion(versionNumber *inventory.VersionNumber, ver inventory.Version) inventory.Versions {
	verB, ok := ver.(*versionBase)
	if !ok {
		panic(fmt.Sprintf("cannot convert to versionBase '%s'", versionNumber))
	}
	v.versions[versionNumber.Int()] = verB
	v.versionInts[versionNumber.String()] = versionNumber.Int()
	return v
}

func (v *versionsBase) IsEmpty() bool {
	return len(v.versions) == 0
}

func (v *versionsBase) GetVersion(versionNumber *inventory.VersionNumber) inventory.Version {
	if versionNumber.IsLatest() {
		versionNumber = v.LatestVersionNumber()
	}
	version, ok := v.versions[versionNumber.Int()]
	if !ok {
		return nil
	}
	return version
}

func (v *versionsBase) Equals(other inventory.Versions) bool {
	otherBase, ok := other.(*versionsBase)
	if !ok {
		return false
	}
	if len(v.versions) != len(otherBase.versions) {
		return false
	}
	for key, version := range v.Iterate() {
		otherVersion := otherBase.GetVersion(key)
		if otherVersion.Err() != nil {
			return false
		}
		return version.Equals(otherVersion)
	}
	return false
}

func (v *versionsBase) String() string {
	var result string
	for k := range v.GetVersionNumbers() {
		result += "; " + k.String()
	}
	return strings.TrimLeft(result, "; ")
}

func (v *versionsBase) Iterate() func(yield func(versionNumber *inventory.VersionNumber, version inventory.Version) bool) {
	return func(yield func(versionString *inventory.VersionNumber, version inventory.Version) bool) {
		for versionNumber := range v.GetVersionNumbers() {
			if !yield(versionNumber, v.versions[versionNumber.Int()]) {
				return
			}
		}
	}
}

func (v *versionsBase) UnmarshalJSON(data []byte) error {
	var newVersions = map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &newVersions); err != nil {
		v.err = errors.Wrapf(err, "cannot unmarshal versions '%s'", string(data))
		return nil
	}
	v.versions = map[int]inventory.Version{}
	for key, bytes := range newVersions {
		versionNumber := inventory.NewVersionNumber().WithString(key)
		ver := v.factory.NewVersion(nil)
		if err := json.Unmarshal(bytes, &ver); err != nil {
			v.err = errors.Wrapf(err, "cannot unmarshal version '%s': '%s'", versionNumber.String(), string(bytes))
			return nil
		}
		ver.WithVersion(versionNumber)
		v.versions[versionNumber.Int()] = ver
		v.versionInts[versionNumber.String()] = versionNumber.Int()
	}
	return nil
}

func (v *versionsBase) MarshalJSON() ([]byte, error) {
	var newVersions = map[string]inventory.Version{}
	for versionNumber := range v.GetVersionNumbers() {
		newVersions[versionNumber.String()] = v.versions[versionNumber.Int()]
	}

	return json.Marshal(newVersions)
}

func (v *versionsBase) LatestVersionNumber() *inventory.VersionNumber {
	if len(v.versions) == 0 {
		return &inventory.VersionNumber{}
	}
	var latest *inventory.VersionNumber
	for v2int, v2 := range v.versions {
		if latest.Int() <= v2int {
			latest = v2.GetVersionNumber()
		}
	}
	return latest
}

var _ inventory.Versions = (*versionsBase)(nil)
