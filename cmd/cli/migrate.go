package main

func doMigrate(arg2, arg3 string) error {
	dsn := getDSN()

	// run the migration command
	switch arg2 {
	case "up":
		err := nap.MigrateUp(dsn)
		if err != nil {
			return err
		}

	case "down":
		if arg3 == "all" {
			err := nap.MigrateDownAll(dsn)
			if err != nil {
				return err
			}
		} else {
			err := nap.Steps(-1, dsn)
			if err != nil {
				return err
			}
		}
	case "reset":
		err := nap.MigrateDownAll(dsn)
		if err != nil {
			return err
		}
		err = nap.MigrateUp(dsn)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
