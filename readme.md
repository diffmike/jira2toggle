### Sync Toggl time entries with Jira worklogs
It takes your jira issues with In Review or In Progress state  
and check them according the latest toggle entries if there are such issues keys in a description,  
if it is found - the time entry log from toggl will be synced into the found jira issue 

#### Configuration
1. Create file `.toggjir` in your home directory  
2. Fill the file with parameters (take a look at an example file)  
2.1 The first line should be filled with full jira url  
2.2 The second line is your jira login  
2.3 The third line is your jira password (can be generated in the [jira cloud account page](https://id.atlassian.com/manage-profile/security/api-tokens))  
2.4 The fourth line is your toggle token (can be found at [the toggle profile page](https://track.toggl.com/profile))  
2.5 (optional) On the fifth line you can put your own Jira Query (the default is   
`assignee = currentUser() and (status = 'In Review' or status = 'In progress') order by updated desc`)

#### Usage
1. Execute the app `./toggjir`
2. Profit