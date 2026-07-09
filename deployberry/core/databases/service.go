package databases

import (
	"deployberry/core/databases/common"
	"shared/repository"
)

func CheckOneService(db string) (common.DBVersion, error) {
	dbConn := repository.GetDB()
	var databaseServer repository.DatabaseServer
	err := dbConn.Where("type = ?", db).First(&databaseServer).Error
	if err != nil {
		return common.DBVersion{}, err
	}

	return common.DBVersion{
		Version:      databaseServer.Version,
		Active:       databaseServer.Active,
		RootPassword: databaseServer.RootPassword,
		Port:         databaseServer.Port,
	}, nil
}
