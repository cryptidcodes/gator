# GATOR CLI TOOL

GATOR is a CLI tool that allows users to:

1. Add RSS feeds from across the internet to be collected
2. Store the collected posts in a PostgreSQL database
3. Follow and unfollow RSS feeds that other users have added
4. View summaries of the aggregated posts in the terminal, with a link to the full post

___

## Installation

The tool uses Go to aggregate feeds into a Postgres SQL database, therefore you will need to install:

1. Go Version 1.22.3 or higher

Linux users can use this command in their terminal:

```
curl -sS https://webi.sh/golang | sh
```

Or use the [official installation instructions](https://go.dev/doc/install)

Confirm that it is installed with this command:

```
go version
```

2. PostgreSQL Version 16.9 or higher

Linux users can use this command:

```
sudo apt update
sudo apt install postgresql postgresql-contrib
```

Confirm that it is installed by running this command:

```
psql --version
```

3. Goose - a database migration tool

Since goose is written in go, we can use this command to install it:

```
go install github.com/pressly/goose/v3/cmd/goose@latest
```

___

## Database Setup

In order to initialize the postgresql database you will need to do the following:

1. Update your postgres password using this command:

```
sudo passwd postgres
```

Ensure it is something you won't forget - you can just make it "postgres"

2. Start the postgres server:

```
sudo service postgresql start
```

3. Connect to the server using psql:

```
sudo -u postgres psql
```

You should see this:

```
postgres=#
```

4. Create a new database. We can call it 'gator'. Use this SQL command:

```
CREATE DATABASE gator;
```

5. Connect to the new database like this:

```
\c gator
```

You should now see this instead:

```
gator=#
```

6. Set the user password for the database. For example:

```
ALTER USER postgres PASSWORD 'postgres';
```

7. Now we can query the database. To test use this query:

```
SELECT version();
```

8. We need to create our connection string to tell Goose how to connect to the gator database. It uses this format:

```
protocol://username:password@host:port/database
```

In our case we are using postgres as the protocol, username, and password. Postgres runs on port 5432 by default and the database is called gator, so our connection string will look like this:

```
postgres://postgres:postgres@localhost:5432/gator
```

You can test it by connecting to the database with psql using the string. If you are already connected to the database type exit to leave, then use this command that should connect you back to the database:

```
psql <connection string>
```

For example:

```
psql postgres://postgres:postgres@localhost:5432/gator
```

If you connect back to the database and see "gator=#" again, then it worked. Exit back out, and move on to migrating the database up to the final structure.

9. Now CD in to the sql/schema directory. If your terminal is already in the gator directory, just use this command:

```
cd sql/schema
```

Otherwise you will have to navigate to the parent folder where gator was installed first. Once the terminal path is set to the schema folder we can run the Goose migration. It will be a command like this:

```
goose postgres <connection string> up
```

Once you get the "database migrated successfully" message, move on to the next step.

10. Lastly, we will need to add the connection string to a config file at the home directory. It will be a json file: ~/.gatorconfig.json

The contents of the file should be a simple json object with the connection string plus a special query at the end to disable SSL like this:

```
protocol://username:password@host:port/database?sslmode=disable
```

The config file should look like this:

```
{
    "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
}
```

Once you have completed this step, you are ready to start using gator!

___

## Commands

Here are the commands available to gator users:

### Users

1. register: registers a new user to the database and logs that user in. Usage:

```
gator register <username>
```

2. login: logs in a registered user. Usage:

```
gator login <username>
```

3. users: prints a list of all users of the database to the console. Usage:

```
gator users
```

### Feeds

1. addfeed: adds a feed source to the database. Usage:

```
gator addfeed <feed name> <feed url>
```

2. feeds: prints a list of feeds in the database to the console. Usage:

```
gator feeds
```

3. follow: adds a feed to the logged in user's following list. Usage:

```
gator follow <feed url>
```

4. following: prints a list of feeds the logged in user is following. Usage:

```
gator following
```

5. unfollow: removes a feed from the logged in user's following list. Usage:

```
gator unfollow <feed url>
```

6. browse: lets the user browse the the most recent posts in their feed. Optionally, set a limit of posts to be shown (default 2 posts). Usage:

```
gator browse <(optional) limit>
```