package main

type DownloadManifest struct {
	packageName string
}

func newDownloadManifest(name string) *DownloadManifest {
	return &DownloadManifest{packageName: name}
}

func (d *DownloadManifest) getManifestURL() string {
	return npmResgistryURL + d.packageName
}

func (d *DownloadManifest) download() error {
	url := d.getManifestURL()
	filename := "manifest/" + d.packageName + ".json"
	return downloadFile(url, filename)
}
