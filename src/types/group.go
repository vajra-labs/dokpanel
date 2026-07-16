package types

type Action string

const (
	Read    Action = "read"
	Create  Action = "create"
	Update  Action = "update"
	Delete  Action = "delete"
	Write   Action = "write"
	Cancel  Action = "cancel"
	Restore Action = "restore"
)

type Statements map[string][]Action

// Master list of all resources and possible actions.
var DefaultStatements = Statements{
	"organization":       {Update, Delete},
	"member":             {Read, Create, Update, Delete},
	"invitation":         {Create, Cancel},
	"project":            {Create, Delete},
	"service":            {Create, Read, Delete},
	"environment":        {Create, Read, Delete},
	"docker":             {Read},
	"sshKeys":            {Read, Create, Delete},
	"gitProviders":       {Read, Create, Delete},
	"traefikFiles":       {Read, Write},
	"volume":             {Read, Create, Delete},
	"deployment":         {Read, Create, Cancel},
	"envVars":            {Read, Write},
	"projectEnvVars":     {Read, Write},
	"environmentEnvVars": {Read, Write},
	"server":             {Read, Create, Delete},
	"registry":           {Read, Create, Delete},
	"certificate":        {Read, Create, Update, Delete},
	"backup":             {Read, Create, Update, Delete, Restore},
	"volumeBackup":       {Read, Create, Update, Delete, Restore},
	"schedule":           {Read, Create, Update, Delete},
	"domain":             {Read, Create, Delete},
	"destination":        {Read, Create, Delete},
	"notification":       {Read, Create, Update, Delete},
	"tag":                {Read, Create, Update, Delete},
	"logs":               {Read},
	"monitoring":         {Read},
	"auditLog":           {Read},
}

// Full access except organization delete.
var AdminStatements = Statements{
	"organization":       {Update},
	"member":             {Read, Create, Update, Delete},
	"invitation":         {Create, Cancel},
	"project":            {Create, Delete},
	"service":            {Create, Read, Delete},
	"environment":        {Create, Read, Delete},
	"docker":             {Read},
	"sshKeys":            {Read, Create, Delete},
	"gitProviders":       {Read, Create, Delete},
	"traefikFiles":       {Read, Write},
	"volume":             {Read, Create, Delete},
	"deployment":         {Read, Create, Cancel},
	"envVars":            {Read, Write},
	"projectEnvVars":     {Read, Write},
	"environmentEnvVars": {Read, Write},
	"server":             {Read, Create, Delete},
	"registry":           {Read, Create, Delete},
	"certificate":        {Read, Create, Update, Delete},
	"backup":             {Read, Create, Update, Delete, Restore},
	"volumeBackup":       {Read, Create, Update, Delete, Restore},
	"schedule":           {Read, Create, Update, Delete},
	"domain":             {Read, Create, Delete},
	"destination":        {Read, Create, Delete},
	"notification":       {Read, Create, Update, Delete},
	"tag":                {Read, Create, Update, Delete},
	"logs":               {Read},
	"monitoring":         {Read},
	"auditLog":           {Read},
}

// Read only on core, limited write on service resources.
var MemberStatements = Statements{
	"organization":       {},
	"member":             {},
	"invitation":         {},
	"project":            {},
	"service":            {Read},
	"environment":        {Read},
	"docker":             {},
	"sshKeys":            {},
	"gitProviders":       {},
	"traefikFiles":       {},
	"volume":             {Read, Create, Delete},
	"deployment":         {Read, Create, Cancel},
	"envVars":            {Read, Write},
	"projectEnvVars":     {Read, Write},
	"environmentEnvVars": {Read, Write},
	"backup":             {Read, Create, Update, Delete, Restore},
	"volumeBackup":       {Read, Create, Update, Delete, Restore},
	"schedule":           {Read, Create, Update, Delete},
	"domain":             {Read, Create, Delete},
	"tag":                {Read},
	"logs":               {Read},
	"monitoring":         {Read},
	"server":             {},
	"registry":           {},
	"certificate":        {},
	"destination":        {},
	"notification":       {},
	"auditLog":           {},
}
