package models

import "time"

type Team struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Slug        string       `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Name        string       `gorm:"size:200;not null" json:"name"`
	Description string       `gorm:"size:500" json:"description,omitempty"`
	CreatedBy   uint         `gorm:"not null" json:"created_by"`
	Creator     User         `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Members     []TeamMember `json:"members,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type TeamRole string

const (
	RoleOwner  TeamRole = "owner"
	RoleEditor TeamRole = "editor"
	RoleViewer TeamRole = "viewer"
)

type TeamMember struct {
	ID     uint     `gorm:"primaryKey" json:"id"`
	TeamID uint     `gorm:"uniqueIndex:idx_team_user;not null" json:"team_id"`
	UserID uint     `gorm:"uniqueIndex:idx_team_user;not null" json:"user_id"`
	Role   TeamRole `gorm:"size:20;not null;default:viewer" json:"role"`
	User   User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Team   Team     `gorm:"foreignKey:TeamID" json:"-"`
}

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationDeclined InvitationStatus = "declined"
)

type TeamInvitation struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	TeamID    uint             `gorm:"index;not null" json:"team_id"`
	UserID    uint             `gorm:"index;not null" json:"user_id"`
	InviterID uint             `gorm:"not null" json:"inviter_id"`
	Role      TeamRole         `gorm:"size:20;not null;default:viewer" json:"role"`
	Status    InvitationStatus `gorm:"size:20;not null;default:pending" json:"status"`
	Team      Team             `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User      User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Inviter   User             `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}
