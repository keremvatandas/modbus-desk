import {profiles} from '../../wailsjs/go/models';

type Props = {
    profiles: profiles.Profile[];
    profileName: string;
    busy: boolean;
    connected: boolean;
    onProfileNameChange: (name: string) => void;
    onSave: () => void;
    onLoad: (profile: profiles.Profile) => void;
    onDelete: (name: string) => void;
};

export function ProfilesPanel({profiles, profileName, busy, connected, onProfileNameChange, onSave, onLoad, onDelete}: Props) {
    return (
        <section className="panel profiles-panel">
            <div className="panel-header">
                <h2>Profiles</h2>
                <span className="count">{profiles.length}</span>
            </div>
            <label>
                Profile name
                <input value={profileName} onChange={(event) => onProfileNameChange(event.target.value)}/>
            </label>
            <button className="primary" disabled={busy || !profileName.trim()} onClick={onSave}>Save profile</button>
            <div className="profile-list">
                {profiles.length === 0 ? (
                    <div className="empty-state compact">No profiles</div>
                ) : profiles.map((profile) => (
                    <div className="profile-row" key={profile.name}>
                        <button className="profile-main" disabled={connected} onClick={() => onLoad(profile)}>
                            <strong>{profile.name}</strong>
                            <span>{profile.host}:{profile.port} / unit {profile.unitId}</span>
                        </button>
                        <button className="icon-button" onClick={() => onDelete(profile.name)} aria-label={`Delete ${profile.name}`}>Delete</button>
                    </div>
                ))}
            </div>
        </section>
    );
}
