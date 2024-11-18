import cv2
import os
import sys
import numpy as np 
import math

## define ROIs
xp=580
yp=399
width = 180
height = 180
xpr=xp+width
ypr=yp+height
mask = None
# Load the cascade
face_cascade = cv2.CascadeClassifier('haarcascade_frontalface_default.xml')

## define mask
def get_mask(width,height):
  mask = np.zeros((height,width,3), np.uint8)
  mask[:,:] = (180,180,180)
  mask[yp:ypr,xp:xpr] = (0,0,0)
  return mask

## set roi
def roi(im):
  global xp
  global yp
  global xpr
  global ypr
  if im is not None:
    im_vis = cv2.addWeighted(im,1.0,mask,0.8,0)
    im_roi = im[yp:ypr,xp:xpr]
    cv2.rectangle(im_vis, (xp, yp), (xpr, ypr), (255,255,255), 2)
    return im_vis,im_roi
  else:
    return None,None

## check brightness level - reject too bright or too dark
def get_brightness(im):
  hsv = cv2.cvtColor(im,cv2.COLOR_BGR2HSV)
  bim = hsv[:,:,2]
  return np.mean(bim,axis=(0,1))

## check blur level - reject blurry images
def get_blurness(im):
  # compute the Laplacian of the image and then return the focus
  # measure, which is simply the variance of the Laplacian
  im_gray = cv2.cvtColor(im,cv2.COLOR_BGR2GRAY)
  return cv2.Laplacian(im_gray, cv2.CV_64F).var()

## check image sharpness
def get_sharpness(im):
  img_HLS = cv2.cvtColor(im, cv2.COLOR_BGR2HLS)
  L = img_HLS[:, :, 1]
  u = np.mean(L)
  LP = cv2.Laplacian(L, cv2.CV_16S, ksize = 3)  
  return np.sum(LP/u) 

## check image contrast
def get_contrast(im):
  sum = 0
  color = ('b','g','r')
  w,h,c = im.shape
  s = w*h
  for i,col in enumerate(color):  
    histogram = cv2.calcHist([im],[i],None,[64],[0,256])
    histogram = np.array(histogram)/s
    for k in range(len(histogram)):
      if histogram[k] != 0:					
        sum = sum - (histogram[k] * math.log(histogram[k]))
  return sum 

## check face
def check_face(im,im_vis):
  im_gray = cv2.cvtColor(im,cv2.COLOR_BGR2GRAY)
  faces = face_cascade.detectMultiScale(im_gray, 1.1, 4)
  if len(faces)>0:
    f = faces[0]
    cv2.rectangle(im_vis, (f[0], f[1]), (f[0]+f[2], f[1]+f[3]), (0,255,0), 2)
    #overlap between face and tongue ROI
    SI = max(0, min(xpr, f[0]+f[2]) - max(xp, f[0])) * max(0, min(ypr, f[1]+f[3]) - max(yp, f[1]))
    #overlap between face-tongue and tongue ROI
    fact = 0.25 
    ft = [f[0]+f[2]*fact,f[1]+f[3]*0.5,f[2]*(1-2*fact),f[3]*0.5]
    ft = [int(i) for i in ft]
    tp = [xp,yp,width,height/2]
    tp = [int(i) for i in tp]
    cv2.rectangle(im_vis, (ft[0], ft[1]), (ft[0]+ft[2], ft[1]+ft[3]), (0,255,255), 2)
    cv2.rectangle(im_vis, (tp[0], tp[1]), (tp[0]+tp[2], tp[1]+tp[3]), (255,255,0), 2)
    ST = max(0, min(tp[0]+tp[2], ft[0]+ft[2]) - max(tp[0], ft[0])) * max(0, min(tp[1]+tp[3], ft[1]+ft[3]) - max(tp[1], ft[1]))
    return SI/(f[2]*f[3]),ST/(tp[2]*tp[3])
  else:
    return 0,0
 
## show texts
def text_on_im(im,xp,yp,text):
  # specify the font and draw the key using puttext
  font = cv2.FONT_HERSHEY_SIMPLEX
  cv2.putText(im,text,(xp,yp), font, 1.2,(82,28,66),2,cv2.LINE_AA)

def write_measure_on_im(im,xp,yp,blur=0,brightness=0, sharpness=0, contrast=0, si=0, st=0, quality=1.0):
  # specify the font and draw the key using puttext
  dis = 20
  s=0.6
  color = (80,0,70)
  font = cv2.FONT_HERSHEY_SIMPLEX
  cv2.putText(im,"blur: "+str("%.2f"%blur),(xp,yp), font, s,color,2,cv2.LINE_AA)
  cv2.putText(im,"brightness: "+str("%.2f"%brightness),(xp,yp+dis), font, s,color,2,cv2.LINE_AA)
  cv2.putText(im,"sharpness: "+str("%.2f"%sharpness),(xp,yp+dis*2), font, s,color,2,cv2.LINE_AA)
  cv2.putText(im,"contrast: "+str("%.2f"%contrast),(xp,yp+dis*3), font, s,color,2,cv2.LINE_AA)
  cv2.putText(im,"tongue/face: "+str("%.2f"%si),(xp,yp+dis*4), font, s,color,2,cv2.LINE_AA)
  cv2.putText(im,"face pos: "+str("%.2f"%st),(xp,yp+dis*5), font, s,color,2,cv2.LINE_AA)
  cv2.putText(im,"quality: "+str("%.2f"%quality),(xp,yp+dis*6), font, s,color,2,cv2.LINE_AA)

## test codes
name = input("What's your name? ")

capture = cv2.VideoCapture(0)

text_dur = 0
text = ""
while(True):   
    ret, frame = capture.read()
    if mask is None:
      h,w,c = frame.shape
      mask = get_mask(w,h)
    frame_vis,frame_roi = roi(frame)
    
    brightness = get_brightness(frame_roi)
    blurness = get_blurness(frame_roi)
    sharpness = get_sharpness(frame_roi)
    contrast = get_contrast(frame_roi) 
    si,st = check_face(frame,frame_vis)  
  
    if brightness>300: 
        text="too bright"
    elif brightness<20:
        text="too dark"
    elif blurness<10:
        text="image is blurry"
    elif si>0.35:
        text="face partially in ROI"
    elif st<0.5:
        text="face not properly placed"
    else:
        text="good quality"
    
    text_on_im(frame_vis,xp-50,yp-30,text)
    write_measure_on_im(frame_vis,0,15,blurness,brightness,sharpness,contrast,si,st)

    cv2.imshow('video', frame_vis)
    c= cv2.waitKey(1) 
    if c== 27: #'esc'
      break
    elif c == ord('p'): #'p'
        while(True):
          cv2.imshow('tongue',frame_roi)
          cv2.imwrite(name+'_tongue.jpg',frame_roi)
          if cv2.waitKey(1) == ord('c'):
            cv2.destroyWindow('tongue')
            break

capture.release()
cv2.destroyAllWindows()
