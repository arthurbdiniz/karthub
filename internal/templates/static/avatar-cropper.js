function avatarCropper() {
    return {
        showModal: false,
        imageSrc: '',
        preview: '',
        croppedData: '',
        cropper: null,
        init() {},
        onFileSelect(e) {
            const file = e.target.files[0];
            if (!file) return;
            if (file.size > 10 * 1024 * 1024) {
                alert('Image is too large (max 10MB). Please choose a smaller file.');
                e.target.value = '';
                return;
            }
            const reader = new FileReader();
            reader.onload = (ev) => {
                this.imageSrc = ev.target.result;
                this.showModal = true;
                this.$nextTick(() => {
                    if (this.cropper) this.cropper.destroy();
                    this.cropper = new Cropper(this.$refs.cropImg, {
                        aspectRatio: 1,
                        viewMode: 2,
                        dragMode: 'move',
                        autoCropArea: 0.9,
                        cropBoxResizable: true,
                        cropBoxMovable: true,
                        background: true,
                        guides: true,
                        center: true,
                        highlight: true,
                        responsive: true,
                    });
                });
            };
            reader.readAsDataURL(file);
        },
        applyCrop() {
            const canvas = this.cropper.getCroppedCanvas({
                width: 200,
                height: 200,
                imageSmoothingQuality: 'high',
            });
            this.croppedData = canvas.toDataURL('image/jpeg', 0.85);
            this.preview = this.croppedData;
            this.showModal = false;
            this.cropper.destroy();
            this.cropper = null;
        },
        cancelCrop() {
            this.showModal = false;
            if (this.cropper) {
                this.cropper.destroy();
                this.cropper = null;
            }
        },
        clearPreview() {
            this.preview = '';
            this.croppedData = '';
        }
    };
}
