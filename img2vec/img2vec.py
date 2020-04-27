import torch
from torchvision import models
from torchvision import transforms

class Image2Vec:

    def __init__(self, embedding_size=512):
        self.device = torch.device(torch.device('cuda') if torch.cuda.is_available() else torch.device('cpu'))

        self.model = models.resnet18(pretrained=True)
        self.embedding_layer = self.model._modules.get('avgpool')
        self.emdedding_size = 512
        
        self.scaler = transforms.Resize((224, 224))
        self.normalize = transforms.Normalize(mean = [0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225])
        self.to_tensor = transforms.ToTensor() 

    OUTPUT_TENSOR = "tensor"
    OUTPUT_VECTOR = "vector"

    def apply_single(self, img, output=OUTPUT_VECTOR):
        image_tensor = self.normalize(self.to_tensor(self.scaler(img))).unsqueeze(0).to(self.device)

        tensor_vec = torch.zeros(1, self.emdedding_size, 1, 1)

        h = self.embedding_layer.register_forward_hook(lambda model, input, output: tensor_vec.copy_(output.data))
        _ = self.model(image_tensor)
        h.remove()

        if output == self.OUTPUT_TENSOR:
            return tensor_vec
        else:
            return tensor_vec.numpy()[0, :, 0, 0]

    def apply_batch(self, imgs, output=OUTPUT_VECTOR):
        images = torch.stack([self.normalize(self.to_tensor(self.scaler(img))) for img in imgs]).to(self.device)

        tensor_vec = torch.zeros(len(imgs), self.emdedding_size, 1, 1)

        h = self.embedding_layer.register_forward_hook(lambda model, input, output: tensor_vec.copy_(output.data))
        _ = self.model(images)
        h.remove()

        if output == self.OUTPUT_TENSOR:
            return tensor_vec
        else:
            return tensor_vec.numpy()[:, :, 0, 0]
